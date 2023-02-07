package bizfly

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	api "github.com/bizflycloud/gobizfly"
	orm "gorm.io/gorm"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
)

type Api interface {
	GetKubeconfig(name string) (string, error)
	GetAccount() string
	GetHost() string
	GetRegion() string
	GetToken() string
	GetUserInfo() (*api.User, error)

	SetRegion(region string) error
	SetToken() error

	ListFirewall() ([]*api.Firewall, error)
	ListCluster() ([]*api.Cluster, error)
	ListServer() ([]*api.Server, error)
	ListVolume() ([]*api.Volume, error)

	SyncFirewall() error
	SyncCluster() error
	SyncServer() error
	SyncVolume() error
}

type apiImpl struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	dbConn     *orm.DB
	token      *api.Token
	client     *api.Client
	host       string
	region     string
	username   string
	password   string
	projectId  string
}

func NewApiFromDatabase(host, region string, timeout time.Duration) ([]Api, error) {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	dbConn.Migrator().CreateTable(
		&AccountModel{},
		&ClusterModel{},
		&ServerModel{},
		&VolumeModel{},
		&FirewallModel{},
		&FirewallBoundModel{},
	)

	ret := make([]Api, 0)
	rows, err := dbConn.Model(&AccountModel{}).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record AccountModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		client, err := newApiWithProjectIdAndUpdateDb(
			host,
			region,
			record.ProjectId,
			record.Email,
			record.Password,
			timeout,
			false,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"Login %s with password %s failt: %v",
				record.Email,
				record.Password,
				err,
			)
		}

		(client.(*apiImpl)).dbConn = dbConn
		ret = append(ret, client)
	}
	return ret, nil
}

func NewApi(
	host string,
	region string,
	username, password string,
	timeout time.Duration,
) (Api, error) {
	return NewApiWithProjectId(
		host, region, "",
		username,
		password,
		timeout,
	)
}

func NewApiWithProjectId(
	host string,
	region string,
	projectId string,
	username, password string,
	timeout time.Duration,
) (Api, error) {
	return newApiWithProjectIdAndUpdateDb(
		host,
		region,
		projectId,
		username, password,
		timeout,
		true,
	)
}

func newApiWithProjectIdAndUpdateDb(
	host string,
	region string,
	projectId string,
	username, password string,
	timeout time.Duration,
	updateDb bool,
) (Api, error) {
	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		time.Millisecond*timeout,
	)
	defer cancelFunc()

	client, err := api.NewClient(api.WithAPIUrl(host), api.WithRegionName(region))
	if err != nil {
		return nil, err
	}

	wrapper := &apiImpl{
		ctx:        ctx,
		host:       host,
		client:     client,
		region:     region,
		cancelFunc: cancelFunc,
		username:   username,
		password:   password,
		projectId:  projectId,
	}

	if updateDb {
		dbModule, err := container.Pick("elephansql")
		if err == nil {
			dbConn, err := db.Establish(dbModule)
			if err != nil {
				return nil, err
			}

			wrapper.dbConn = dbConn
		}
	}

	err = wrapper.SetToken()
	if err != nil {
		return nil, err
	}

	return wrapper, nil
}

func (self *apiImpl) SetToken() error {
	token, err := self.client.Token.Create(
		self.ctx,
		&api.TokenCreateRequest{
			AuthMethod: "password",
			Username:   self.username,
			Password:   self.password,
			ProjectID:  self.projectId,
		},
	)
	if err != nil {
		return err
	} else {
		self.token = token
	}

	self.client.SetKeystoneToken(self.token)

	if self.dbConn != nil {
		user, err := self.GetUserInfo()
		if err != nil {
			return err
		}

		projectId := "default"

		if len(self.projectId) > 0 {
			projectId = self.projectId
		}

		if len(user.BillingAccID) > 0 {
			result := self.dbConn.FirstOrCreate(
				&AccountModel{
					BaseModel: BaseModel{
						UUID: fmt.Sprintf("%s-%s", user.BillingAccID, projectId),
					},
					Email:     self.username,
					Password:  self.password,
					ProjectId: self.projectId,
				},
				AccountModel{
					BaseModel: BaseModel{
						UUID: fmt.Sprintf("%s-%s", user.BillingAccID, projectId),
					},
				},
			)
			return result.Error
		} else {
			result := self.dbConn.FirstOrCreate(
				&AccountModel{
					BaseModel: BaseModel{
						UUID: fmt.Sprintf("fakeid:%s-%s", self.username, projectId),
					},
					Email:     self.username,
					Password:  self.password,
					ProjectId: self.projectId,
				},
				AccountModel{
					BaseModel: BaseModel{
						UUID: fmt.Sprintf("fakeid:%s-%s", self.username, projectId),
					},
				},
			)
			return result.Error
		}
	}
	return nil
}

func (self *apiImpl) SetRegion(region string) error {
	err := api.WithRegionName(region)(self.client)
	if err != nil {
		return err
	}

	return self.SetToken()
}

func (self *apiImpl) GetAccount() string {
	return self.username
}

func (self *apiImpl) GetHost() string {
	return self.host
}

func (self *apiImpl) GetToken() string {
	return self.token.KeystoneToken
}

func (self *apiImpl) GetRegion() string {
	return self.region
}

func (self *apiImpl) GetUserInfo() (*api.User, error) {
	defer self.cancelFunc()

	user, err := callBizflyApiWithMeasurement(
		"get-user-info",
		func() (interface{}, error) {
			return self.client.Account.GetUserInfo(self.ctx)
		},
	)

	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return nil, errors.New(msg)
	}

	return user.(*api.User), nil
}

func (self *apiImpl) GetKubeconfig(clusterId string) (string, error) {
	defer self.cancelFunc()

	return self.client.KubernetesEngine.GetKubeConfig(
		self.ctx,
		clusterId,
	)
}

func (self *apiImpl) ListCluster() ([]*api.Cluster, error) {
	clusters := make([]*api.Cluster, 0)
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	rows, err := dbConn.Model(&ClusterModel{}).
		Where("account = ?", self.GetAccount()).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record ClusterModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, &api.Cluster{
			UID:             record.UUID,
			Name:            record.Name,
			ProvisionStatus: record.Status,
		})
	}

	return clusters, nil
}

func (self *apiImpl) ListServer() ([]*api.Server, error) {
	servers := make([]*api.Server, 0)
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	rows, err := dbConn.Model(&ServerModel{}).
		Where("account = ?", self.GetAccount()).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record ServerModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		servers = append(servers, &api.Server{
			ID:     record.UUID,
			Status: record.Status,
		})
	}

	return servers, nil
}

func (self *apiImpl) ListFirewall() ([]*api.Firewall, error) {
	firewalls := make([]*api.Firewall, 0)
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	rows, err := dbConn.Model(&FirewallModel{}).
		Where("account = ?", self.GetAccount()).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var firewallRecord FirewallModel

		inbound := make([]api.FirewallRule, 0)
		outbound := make([]api.FirewallRule, 0)

		err = dbConn.ScanRows(rows, &firewallRecord)
		if err != nil {
			return nil, err
		}

		boundRows, err := dbConn.Model(&FirewallBoundModel{}).
			Where("firewall = ?", firewallRecord.UUID).
			Rows()

		err = dbConn.ScanRows(boundRows, &FirewallBoundModel{})
		if err != nil {
			return nil, err
		}

		for boundRows.Next() {
			var firewallBoundRecord FirewallBoundModel

			err = dbConn.ScanRows(rows, &firewallBoundRecord)
			if err != nil {
				return nil, err
			}

			switch firewallBoundRecord.Type {
			case InBound:
				inbound = append(inbound, api.FirewallRule{
					ID:   firewallRecord.UUID,
					Type: "inbound",
					CIDR: firewallBoundRecord.CIDR,
				})

			case OutBound:
				outbound = append(outbound, api.FirewallRule{
					ID:   firewallRecord.UUID,
					Type: "outbound",
					CIDR: firewallBoundRecord.CIDR,
				})

			default:
				continue
			}
		}

		firewalls = append(firewalls, &api.Firewall{
			BaseFirewall: api.BaseFirewall{
				ID:       firewallRecord.UUID,
				InBound:  inbound,
				OutBound: outbound,
			},
		})
	}

	return firewalls, nil
}

func (self *apiImpl) ListVolume() ([]*api.Volume, error) {
	volumes := make([]*api.Volume, 0)
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	rows, err := dbConn.Model(&VolumeModel{}).
		Where("account = ?", self.GetAccount()).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record VolumeModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		volumes = append(volumes, &api.Volume{
			ID:         record.UUID,
			Status:     record.Status,
			VolumeType: record.Type,
		})
	}

	return volumes, nil
}

func (self *apiImpl) SyncCluster() error {
	clusters, err := callBizflyApiWithMeasurement(
		"list-kubernertes-engine",
		func() (interface{}, error) {
			return self.client.KubernetesEngine.List(self.ctx, nil)
		},
	)

	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return errors.New(msg)
	}

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)

	if err != nil {
		return err
	}

	clusterRecords := make([]ClusterModel, 0)

	for _, cluster := range clusters.([]*api.Cluster) {
		clusterRecords = append(clusterRecords, ClusterModel{
			BaseModel: BaseModel{
				UUID: cluster.UID,
			},
			Account: self.GetAccount(),
			Name:    cluster.Name,
			Status:  cluster.ProvisionStatus,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.Model(ClusterModel{}).
		Where("account = ?", self.GetAccount()).
		Update("status", "Unknown")
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.CreateInBatches(clusterRecords, batchSize)
	return resp.Error
}

func (self *apiImpl) SyncServer() error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	servers, err := callBizflyApiWithMeasurement(
		"list-server",
		func() (interface{}, error) {
			return self.client.Server.List(self.ctx, &api.ServerListOptions{})
		},
	)

	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return errors.New(msg)
	}

	serverRecords := make([]ServerModel, 0)

	for _, server := range servers.([]*api.Server) {
		serverRecords = append(serverRecords, ServerModel{
			BaseModel: BaseModel{
				UUID: server.ID,
			},
			Account: self.GetAccount(),
			Status:  server.Status,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.CreateInBatches(serverRecords, batchSize)
	return resp.Error
}

func (self *apiImpl) SyncFirewall() error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	firewalls, err := callBizflyApiWithMeasurement(
		"list-firewall",
		func() (interface{}, error) {
			return self.client.Firewall.List(self.ctx, nil)
		},
	)

	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return errors.New(msg)
	}

	firewallRecords := make([]FirewallModel, 0)
	firewallBoundRecords := make([]FirewallBoundModel, 0)

	for _, firewall := range firewalls.([]*api.Firewall) {
		firewallRecords = append(firewallRecords, FirewallModel{
			BaseModel: BaseModel{
				UUID: firewall.ID,
			},
			Account: self.GetAccount(),
		})

		for _, bound := range firewall.InBound {
			firewallBoundRecords = append(firewallBoundRecords, FirewallBoundModel{
				BaseModel: BaseModel{
					UUID: bound.ID,
				},
				Firewall: firewall.ID,
				CIDR:     bound.CIDR,
				Type:     0,
			})
		}

		for _, bound := range firewall.OutBound {
			firewallBoundRecords = append(firewallBoundRecords, FirewallBoundModel{
				BaseModel: BaseModel{
					UUID: bound.ID,
				},
				Firewall: firewall.ID,
				CIDR:     bound.CIDR,
				Type:     1,
			})
		}
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.CreateInBatches(firewallRecords, batchSize)
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.CreateInBatches(firewallBoundRecords, batchSize)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (self *apiImpl) SyncVolume() error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	volumes, err := callBizflyApiWithMeasurement(
		"list-kubernertes-engine",
		func() (interface{}, error) {
			return self.client.Volume.List(self.ctx, nil)
		},
	)

	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return errors.New(msg)
	}

	volumeRecords := make([]VolumeModel, 0)

	for _, volume := range volumes.([]*api.Volume) {
		volumeRecords = append(volumeRecords, VolumeModel{
			BaseModel: BaseModel{
				UUID: volume.ID,
			},
			Account: self.GetAccount(),
			Type:    volume.VolumeType,
			Status:  volume.Status,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.CreateInBatches(volumeRecords, batchSize)
	return resp.Error
}
