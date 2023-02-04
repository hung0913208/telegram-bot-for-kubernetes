package bizfly

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/bizflycloud/gobizfly"
	orm "gorm.io/gorm"

    "github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
    "github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
)

type Api interface {
	GetKubeconfig(name string) (string, error)
	GetAccount() string
	GetRegion() string
	GetToken() string
	GetUserInfo() (*api.User, error)

	SetRegion(region string) error
	SetToken() error

	ListFirewall() ([]*api.Firewall, error)
	ListCluster() ([]*api.Cluster, error)
	ListServer() ([]*api.Server, error)
	ListVolume() ([]*api.Volume, error)
}

type apiImpl struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
    dbConn     *orm.DB
	token      *api.Token
	client     *api.Client
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

    ret := make([]Api, 0)
    rows, err := dbConn.Model(&AccountModel{}).
                        Rows()
    if err != nil {
        return nil, err
    }

    for rows.Next() {
        var record AccountModel

        err = dbConn.ScanRows(rows, &record)
        if err != nil {
            return nil, err
        }
       
        client, err := newApiWithProjectIdAndUpdateDb(
            host,
            region,
            record.Email,
            record.Password,
            record.ProjectId,
            timeout,
            false,
        )
        if err != nil {
            return nil, err
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
            result := self.dbConn.FirstOrCreate(&AccountModel{
                UUID:      fmt.Sprintf("%s-%s", user.BillingAccID, projectId),
                Email:     self.username,
                Password:  self.password,
                ProjectId: self.projectId,
            }, 
            AccountModel{ UUID: "non_existing" })
            return result.Error
        } else {
            result := self.dbConn.FirstOrCreate(&AccountModel{
                UUID:      fmt.Sprintf("fakeid:%s-%s", self.username, projectId),
                Email:     self.username,
                Password:  self.password,
                ProjectId: self.projectId,
            },
            AccountModel{ UUID: "non_existing" })
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

		return nil, errors.New(msg)
	}

	return clusters.([]*api.Cluster), nil
}

func (self *apiImpl) ListServer() ([]*api.Server, error) {
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

		return nil, errors.New(msg)
	}

	return servers.([]*api.Server), nil
}

func (self *apiImpl) ListFirewall() ([]*api.Firewall, error) {
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

		return nil, errors.New(msg)
	}

	return firewalls.([]*api.Firewall), nil
}

func (self *apiImpl) ListVolume() ([]*api.Volume, error) {
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

		return nil, errors.New(msg)
	}

	return volumes.([]*api.Volume), nil
}
