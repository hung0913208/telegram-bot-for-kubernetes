package cluster

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

func (self *clusterImpl) getListTenantFromDb() ([]string, error) {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	rows, err := dbConn.Model(&ClusterModel{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := make([]string, 0)
	for rows.Next() {
		var record ClusterModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		if len(record.Name) == 0 {
			break
		}

		tenants = append(tenants, record.Name)
	}

	return tenants, nil
}

func (self *clusterImpl) loadTenantFromDb(name string) error {
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
		var metadata interface{}

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return err
		}

		if len(record.Name) == 0 {
			break
		}

		kubeconfig, err := base64.StdEncoding.DecodeString(record.Kubeconfig)
		if err != nil {
			continue
		}

		tenant, err := kubernetes.NewDefaultTenant(
			record.Name,
			kubeconfig,
		)
		if err != nil {
			return fmt.Errorf("new tenant %s fails: %v", record.Name, err)
		}

		err = json.Unmarshal([]byte(record.Metadata), &metadata)
		if err != nil {
			return err
		}

		// @TODO: seem to be we must remove this one or only use when we face
		//        outdated
		if false {
			tenant, err = record.Provider.ConvertMetadataToTenant(
				record.Metadata,
				self.timeout,
			)
		}
		if err != nil {
			return err
		}

		self.tenants[record.Name] = tenant
	}

	return nil
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

		metadata, err := tenant.GetMetadata()
		if err != nil {
			return err
		}

		encodedMetadata, err := json.Marshal(metadata)
		if err != nil {
			return err
		}

		encodedKubeconfig := []byte(base64.StdEncoding.EncodeToString(
			[]byte(kubeconfig),
		))
		dbConn.FirstOrCreate(
			&ClusterModel{
				Name:       tenant.GetName(),
				Provider:   ProviderEnum(provider),
				Metadata:   string(encodedMetadata),
				Kubeconfig: string(encodedKubeconfig),
			},
			ClusterModel{Kubeconfig: string(encodedKubeconfig)},
		)
		return nil
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
			return nil, fmt.Errorf("Can't find %s", name)
		}
	}

	return tenant, nil
}

func List(module container.Module) ([]string, error) {
	clusterMgr, ok := module.(*clusterImpl)
	if !ok {
		return nil, errors.New("Unknown module")
	}

	return clusterMgr.getListTenantFromDb()
}
