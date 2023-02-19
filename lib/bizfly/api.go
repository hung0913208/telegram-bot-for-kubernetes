package bizfly

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	api "github.com/bizflycloud/gobizfly"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
)

type Api interface {
	GetKubeconfig(name string) (string, error)
	GetAccount() string
	GetHost() string
	GetRegion() string
	GetToken() string
	GetProjectId() string
	GetPool(poolId string) (*PoolModel, error)
	GetUserInfo() (*api.User, error)

	SetRegion(region string) error
	SetToken() error

	ListFirewall() ([]*api.Firewall, error)
	ListCluster() ([]*api.Cluster, error)
	ListServer(clusterId ...string) ([]*api.Server, error)
	ListVolume(serverId ...string) ([]*api.Volume, error)
	ListPool(clusterId string) ([]*api.ExtendedWorkerPool, error)

	SyncFirewall() error
	SyncCluster() error
	SyncServer() error
	SyncVolume() error
	SyncPool(clusterId string) error
	SyncPoolNode(clusterId, poolId string) error
	SyncVolumeAttachment(serverId string) error

	DetachCluster(clusterId string) error
	DetachServer(serverId string) error
	DetachPool(poolId string) error

	LinkPodWithVolume(pod, cluster, volumeId string, size int) error

	// AdjustVolume() error
	// AdjustPool() error
	// AdjustAlert() error

	Clean() error
}

type apiImpl struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	dbConn     *gorm.DB
	token      *api.Token
	client     *api.Client
	uuid       string
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
		&ClusterStatModel{},
		&PoolModel{},
		&PoolNodeModel{},
		&ServerModel{},
		&VolumeModel{},
		&VolumeClusterModel{},
		&VolumeServerModel{},
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

		projectId := ""
		if record.ProjectId != "default" {
			projectId = record.ProjectId
		}

		client, err := newApiWithProjectIdAndUpdateDb(
			host,
			region,
			projectId,
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

		(client.(*apiImpl)).uuid = record.UUID
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
			self.uuid = fmt.Sprintf("%s-%s", user.BillingAccID, projectId)
		} else {
			self.uuid = fmt.Sprintf("fakeid:%s-%s", self.username, projectId)
		}

		result := self.dbConn.FirstOrCreate(
			&AccountModel{
				BaseModel: BaseModel{
					UUID: self.uuid,
				},
				Email:     self.username,
				Password:  self.password,
				ProjectId: projectId,
			},
			AccountModel{
				BaseModel: BaseModel{
					UUID: self.uuid,
				},
			},
		)
		return result.Error
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

func (self *apiImpl) GetProjectId() string {
	return self.projectId
}

func (self *apiImpl) GetPool(poolId string) (*PoolModel, error) {
	var record PoolModel

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	resp := dbConn.First(&record, "uuid = ?", poolId)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return &record, nil
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

	kubeconfig, err := self.client.KubernetesEngine.GetKubeConfig(
		self.ctx,
		clusterId,
	)

	return kubeconfig, err
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
		Where("account = ?", self.uuid).
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
			Tags:            strings.Split(record.Tags, ","),
		})
	}

	return clusters, nil
}

func (self *apiImpl) ListServer(clusterId ...string) ([]*api.Server, error) {
	var rows *sql.Rows

	servers := make([]*api.Server, 0)
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	if len(clusterId) > 0 {
		rows, err = dbConn.Model(&ServerModel{}).
			Where("account = ? and cluster = ?", self.uuid, clusterId[0]).
			Rows()
	} else {
		rows, err = dbConn.Model(&ServerModel{}).
			Where("account = ?", self.uuid).
			Rows()
	}
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
		Where("account = ?", self.uuid).
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

func (self *apiImpl) ListVolume(serverId ...string) ([]*api.Volume, error) {
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
		Where("account = ?", self.uuid).
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

func (self *apiImpl) ListPool(clusterId string) ([]*api.ExtendedWorkerPool, error) {
	var rows *sql.Rows

	pools := make([]*api.ExtendedWorkerPool, 0)
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	if len(clusterId) > 0 {
		rows, err = dbConn.Model(&PoolModel{}).
			Where("cluster = ?", clusterId).
			Rows()
	} else {
		rows, err = dbConn.Model(&PoolModel{}).Rows()
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var record PoolModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		pools = append(pools, &api.ExtendedWorkerPool{
			UID: record.UUID,
			WorkerPool: api.WorkerPool{
				Name:              record.Name,
				EnableAutoScaling: record.EnableAutoscaling,
				MinSize:           record.RequiredSize,
				MaxSize:           record.LimitSize,
			},
			ProvisionStatus:    record.Status,
			AutoScalingGroupID: record.Autoscale,
		})
	}

	return pools, nil
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
			Account: self.uuid,
			Name:    cluster.Name,
			Status:  cluster.ProvisionStatus,
			Locked:  true,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(clusterRecords, batchSize)
	return resp.Error
}

func (self *apiImpl) SyncServer() error {
	dbModule, err := container.Pick("elephansql")
	updateVolumes := true

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
		if server.CreatedAt[len(server.CreatedAt)-1] != 'Z' {
			server.CreatedAt += "Z"
		}

		if server.UpdatedAt[len(server.UpdatedAt)-1] != 'Z' {
			server.UpdatedAt += "Z"
		}

		createdTime, err := time.Parse(time.RFC3339Nano, server.CreatedAt)
		if err != nil {
			return err
		}

		updatedTime, err := time.Parse(time.RFC3339Nano, server.UpdatedAt)
		if err != nil {
			return err
		}

		serverRecords = append(serverRecords, ServerModel{
			BaseModel: BaseModel{
				UUID:      server.ID,
				CreatedAt: createdTime,
				UpdatedAt: updatedTime,
			},
			Account: self.uuid,
			Status:  server.Status,
			Locked:  true,
			Zone:    server.AvailabilityZone,
		})

		if updateVolumes {
			err = self.updateVolumeAttachment(server)

			if err != nil {
				updateVolumes = false
			}
		}
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uuid"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
		}).
		CreateInBatches(serverRecords, batchSize)
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
		if firewall.CreatedAt[len(firewall.CreatedAt)-1] != 'Z' {
			firewall.CreatedAt += "Z"
		}

		if firewall.UpdatedAt[len(firewall.UpdatedAt)-1] != 'Z' {
			firewall.UpdatedAt += "Z"
		}

		createdTime, err := time.Parse(time.RFC3339Nano, firewall.CreatedAt)
		if err != nil {
			return err
		}

		updatedTime, err := time.Parse(time.RFC3339Nano, firewall.UpdatedAt)
		if err != nil {
			return err
		}

		firewallRecords = append(firewallRecords, FirewallModel{
			BaseModel: BaseModel{
				UUID:      firewall.ID,
				CreatedAt: createdTime,
				UpdatedAt: updatedTime,
			},
			Account: self.uuid,
		})

		for _, bound := range firewall.InBound {
			if bound.CreatedAt[len(bound.CreatedAt)-1] != 'Z' {
				bound.CreatedAt += "Z"
			}

			if bound.UpdatedAt[len(bound.UpdatedAt)-1] != 'Z' {
				bound.UpdatedAt += "Z"
			}

			createdTime, err := time.Parse(time.RFC3339Nano, bound.CreatedAt)
			if err != nil {
				return err
			}

			updatedTime, err := time.Parse(time.RFC3339Nano, bound.UpdatedAt)
			if err != nil {
				return err
			}

			firewallBoundRecords = append(firewallBoundRecords, FirewallBoundModel{
				BaseModel: BaseModel{
					UUID:      bound.ID,
					CreatedAt: createdTime,
					UpdatedAt: updatedTime,
				},
				Firewall: firewall.ID,
				CIDR:     bound.CIDR,
				Type:     0,
			})
		}

		for _, bound := range firewall.OutBound {
			if bound.CreatedAt[len(bound.CreatedAt)-1] != 'Z' {
				bound.CreatedAt += "Z"
			}

			if bound.UpdatedAt[len(bound.UpdatedAt)-1] != 'Z' {
				bound.UpdatedAt += "Z"
			}

			createdTime, err := time.Parse(time.RFC3339Nano, bound.CreatedAt)
			if err != nil {
				return err
			}

			updatedTime, err := time.Parse(time.RFC3339Nano, bound.UpdatedAt)
			if err != nil {
				return err
			}

			firewallBoundRecords = append(firewallBoundRecords, FirewallBoundModel{
				BaseModel: BaseModel{
					UUID:      bound.ID,
					CreatedAt: createdTime,
					UpdatedAt: updatedTime,
				},
				Account:  self.uuid,
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

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(firewallRecords, batchSize)
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(firewallBoundRecords, batchSize)
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
		if volume.CreatedAt[len(volume.CreatedAt)-1] != 'Z' {
			volume.CreatedAt += "Z"
		}

		if volume.UpdatedAt[len(volume.UpdatedAt)-1] != 'Z' {
			volume.UpdatedAt += "Z"
		}

		createdTime, err := time.Parse(time.RFC3339Nano, volume.CreatedAt)
		if err != nil {
			return err
		}

		updatedTime, err := time.Parse(time.RFC3339Nano, volume.UpdatedAt)
		if err != nil {
			return err
		}

		volumeRecords = append(volumeRecords, VolumeModel{
			BaseModel: BaseModel{
				UUID:      volume.ID,
				CreatedAt: createdTime,
				UpdatedAt: updatedTime,
			},
			Account:     self.uuid,
			Type:        volume.VolumeType,
			Status:      volume.Status,
			Zone:        volume.AvailabilityZone,
			Description: volume.Description,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uuid"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
		}).
		CreateInBatches(volumeRecords, batchSize)
	return resp.Error
}

func (self *apiImpl) SyncPool(clusterId string) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return errors.New(msg)
	}

	cluster, err := callBizflyApiWithMeasurement(
		"list-pool",
		func() (interface{}, error) {
			return self.client.KubernetesEngine.Get(self.ctx, clusterId)
		},
	)
	if err != nil {
		return err
	}

	poolRecords := make([]PoolModel, 0)

	resp := dbConn.FirstOrCreate(
		&ClusterStatModel{
			Cluster: clusterId,
			Account: self.uuid,
			Core:    cluster.(*api.FullCluster).Stat.TotalCPU,
			Memory:  cluster.(*api.FullCluster).Stat.TotalMemory,
		},
		ClusterStatModel{
			Core:   cluster.(*api.FullCluster).Stat.TotalCPU,
			Memory: cluster.(*api.FullCluster).Stat.TotalMemory,
		},
	)
	if resp.Error != nil {
		return resp.Error
	}

	for _, pool := range cluster.(*api.FullCluster).WorkerPools {
		if pool.CreatedAt[len(pool.CreatedAt)-1] != 'Z' {
			pool.CreatedAt += "Z"
		}

		createdTime, err := time.Parse(time.RFC3339Nano, pool.CreatedAt)
		if err != nil {
			return err
		}

		poolRecords = append(poolRecords, PoolModel{
			BaseModel: BaseModel{
				UUID:      pool.UID,
				CreatedAt: createdTime,
			},
			Cluster:           clusterId,
			Name:              pool.Name,
			Zone:              pool.AvailabilityZone,
			Status:            pool.ProvisionStatus,
			Autoscale:         pool.AutoScalingGroupID,
			EnableAutoscaling: pool.EnableAutoScaling,
			RequiredSize:      pool.MaxSize,
			LimitSize:         pool.MinSize,
			Account:           self.uuid,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp = dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(poolRecords, batchSize)
	return resp.Error
}

func (self *apiImpl) SyncPoolNode(clusterId, poolId string) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	nodes, err := callBizflyApiWithMeasurement(
		"list-pool-nodes",
		func() (interface{}, error) {
			return self.client.KubernetesEngine.GetClusterWorkerPool(
				self.ctx, clusterId, poolId)
		},
	)
	if err != nil {
		msg, bug := removeSvgBlock(fmt.Sprintf("%v", err))
		if bug != nil {
			panic(bug)
		}

		return errors.New(msg)
	}

	nodeRecords := make([]PoolNodeModel, 0)

	for _, node := range nodes.(*api.WorkerPoolWithNodes).Nodes {
		nodeRecords = append(nodeRecords, PoolNodeModel{
			BaseModel: BaseModel{
				UUID: node.ID,
			},
			Name:    node.Name,
			Server:  node.PhysicalID,
			Status:  node.Status,
			Reason:  node.StatusReason,
			Account: self.uuid,
			Cluster: clusterId,
			Pool:    poolId,
		})

		resp := dbConn.Model(&ServerModel{}).
			Where("uuid = ?", node.PhysicalID).
			Update("cluster", clusterId)
		if resp.Error != nil {
			return resp.Error
		}
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(nodeRecords, batchSize)
	return resp.Error
}

func (self *apiImpl) SyncVolumeAttachment(serverId string) error {
	server, err := callBizflyApiWithMeasurement(
		"list-attached-volumes",
		func() (interface{}, error) {
			return self.client.Server.Get(self.ctx, serverId)
		},
	)
	if err != nil {
		return err
	}

	if _, ok := server.(*api.Server); !ok {
		return errors.New("get unknown object")
	}

	return self.updateVolumeAttachment(server.(*api.Server))
}

func (self *apiImpl) LinkPodWithVolume(
	pod, cluster, volumeId string, size int,
) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	resp := dbConn.FirstOrCreate(
		&VolumeClusterModel{
			Volume:  volumeId,
			Account: self.uuid,
			Pod:     pod,
			Cluster: cluster,
			Size:    size,
		},
		VolumeClusterModel{
			Pod:     pod,
			Cluster: cluster,
			Size:    size,
		},
	)
	return resp.Error
}

func (self *apiImpl) DetachCluster(clusterId string) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	resp := dbConn.Where("account = ? and cluster = ?", self.uuid, clusterId).
		Delete(&VolumeClusterModel{Account: self.uuid, Cluster: clusterId})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ? and cluster = ?", self.uuid, clusterId).
		Delete(&PoolNodeModel{Account: self.uuid, Cluster: clusterId})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ? and cluster = ?", self.uuid, clusterId).
		Delete(&ServerModel{Account: self.uuid, Cluster: clusterId})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Delete(&ClusterModel{
		BaseModel: BaseModel{
			UUID: clusterId,
		},
	})
	return resp.Error
}

func (self *apiImpl) DetachPool(poolId string) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	rows, err := dbConn.Model(&PoolNodeModel{}).
		Where("pool = ?", poolId).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var record PoolNodeModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return err
		}

		err = self.DetachServer(record.Server)
		if err != nil {
			return err
		}
	}

	resp := dbConn.Where("account = ? and pool = ?", self.uuid, poolId).
		Delete(&PoolNodeModel{Account: self.uuid})
	return resp.Error
}

func (self *apiImpl) DetachServer(serverId string) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	resp := dbConn.Delete(&ServerModel{
		BaseModel: BaseModel{
			UUID: serverId,
		},
	})
	return resp.Error
}

func (self *apiImpl) Clean() error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	resp := dbConn.Where("account = ?", self.uuid).
		Delete(&ClusterModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&PoolModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&PoolNodeModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&ServerModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&VolumeModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&VolumeClusterModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&VolumeServerModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Where("account = ?", self.uuid).
		Delete(&FirewallModel{Account: self.uuid})
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.
		Where("account = ?", self.uuid).
		Delete(&FirewallBoundModel{Account: self.uuid})
	return resp.Error
}

func (self *apiImpl) updateVolumeAttachment(server *api.Server) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	volumeServerRecords := make([]VolumeServerModel, 0)

	for _, volume := range server.AttachedVolumes {
		volumeServerRecords = append(volumeServerRecords,
			VolumeServerModel{
				Volume:  volume.ID,
				Server:  server.ID,
				Account: self.uuid,
			},
		)
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(volumeServerRecords, batchSize)
	return resp.Error
}
