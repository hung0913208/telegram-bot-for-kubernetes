package cluster

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
)

type Cluster interface {
	container.Module
}

type clusterImpl struct {
	tenants map[string]kubernetes.Tenant
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
	)

	return &clusterImpl{
		tenants: make(map[string]kubernetes.Tenant),
	}, nil
}

func (self *clusterImpl) Init(timeout time.Duration) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	rows, err := dbConn.Model(&ClusterModel{}).Rows()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var record ClusterModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return err
		}

		tenant, err := kubernetes.NewDefaultTenant(
			record.Name,
			record.Kubeconfig,
		)
		if err != nil {
			var metadata interface{}

			err = json.Unmarshal([]byte(record.Metadata), &metadata)
			if err != nil {
				return err
			}

			tenant, err = record.Provider.ConvertMetadataToTenant(
				record.Metadata,
				timeout,
			)
			if err != nil {
				return err
			}
		}

		self.tenants[record.Name] = tenant
	}
	return nil
}

func (self *clusterImpl) Deinit() error {
	return nil
}

func (self *clusterImpl) Execute(args []string) error {
	return errors.New("Don't support using as interactive module")
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
		clusterManager.tenants[tenant.GetName()] = tenant

		provider, err := tenant.GetProvider()
		if err != nil {
			return err
		}

		kubeconfig, err := tenant.GetKubeconfig()
		if err != nil {
			return err
		}

		dbConn.FirstOrCreate(
			&ClusterModel{
				Name:       tenant.GetName(),
				Provider:   ProviderEnum(provider),
				Kubeconfig: kubeconfig,
			},
			ClusterModel{Kubeconfig: kubeconfig},
		)
		return nil
	}
}
