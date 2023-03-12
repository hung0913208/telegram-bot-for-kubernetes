package cluster

import (
	"errors"
	"fmt"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/platform"
)

type Cluster interface {
	container.Module
}

type clusterImpl struct {
	tenants map[string]kubernetes.Tenant
	timeout time.Duration
}

func NewModule() (Cluster, error) {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	dbConn.Migrator().CreateTable(
		&ClusterModel{},
		&AliasModel{},
	)

	return &clusterImpl{
		tenants: make(map[string]kubernetes.Tenant),
	}, nil
}

func (self *clusterImpl) Init(timeout time.Duration) error {
	self.timeout = timeout
	return nil
}

func (self *clusterImpl) Deinit() error {
	return nil
}

func (self *clusterImpl) Execute(args []string) error {
	return errors.New("Don't support using as interactive module")
}

func Detach(clusterName string, module ...string) error {
	moduleName := "cluster"
	if len(module) > 0 {
		moduleName = module[0]
	}

	clusterModule, err := container.Pick(moduleName)
	if err != nil {
		return err
	}

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	if clusterManager, ok := clusterModule.(*clusterImpl); !ok {
		return errors.New("Cannot get module `cluster`")
	} else {
		resp := dbConn.Where("name = ?", clusterName).
			Delete(&ClusterModel{
				Name: clusterName,
			})

		if resp.Error == nil {
			delete(clusterManager.tenants, clusterName)
		}

		return resp.Error
	}
}

func Join(tenant kubernetes.Tenant, module ...string) error {
	moduleName := "cluster"
	if len(module) > 0 {
		moduleName = module[0]
	}

	clusterModule, err := container.Pick(moduleName)
	if err != nil {
		return err
	}

	if clusterMgr, ok := clusterModule.(*clusterImpl); !ok {
		return errors.New("Cannot get module `cluster`")
	} else {
		return clusterMgr.updateTenantToDb(tenant)
	}
}

func Pick(module container.Module, name string) (kubernetes.Tenant, error) {
	clusterMgr, ok := module.(*clusterImpl)
	if !ok {
		return nil, errors.New("Unknown module")
	}

	tenant, ok := clusterMgr.tenants[name]
	if !ok {
		if err := clusterMgr.loadTenantFromDb(name); err != nil {
			return nil, err
		}

		if tenant, ok = clusterMgr.tenants[name]; !ok {
			tenant, err := clusterMgr.convertAliasToTenant(name)
			if err != nil {
				return nil, fmt.Errorf("Can't find %s: %v", name, err)
			}

			return tenant, nil
		}
	}

	return tenant, nil
}

func List(module container.Module) (map[string][]string, error) {
	clusterMgr, ok := module.(*clusterImpl)
	if !ok {
		return nil, errors.New("Unknown module")
	}

	mapping := make(map[string][]string)

	tenants, err := clusterMgr.getListTenantFromDb()
	if err != nil {
		return nil, err
	}

	for _, tenant := range tenants {
		aliases, err := clusterMgr.getListAliasesFromDb(tenant)
		if err != nil {
			return nil, err
		}
		mapping[tenant] = aliases
	}

	return mapping, nil
}

func Scan(module container.Module, cluster string) error {
	tenant, err := Pick(module, cluster)
	if err != nil {
		return err
	}

	client, err := tenant.GetClient()
	if err != nil {
		return err
	}

	pods, err := client.GetPods("")
	if err != nil {
		return err
	}

	_, err = platform.GetPgFromPodList(client, pods)
	if err != nil {
		return err
	}
	return nil
}
