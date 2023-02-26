package cluster

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm/clause"

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
		&NodeModel{},
		&PodModel{},
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

func (self *clusterImpl) getListAliasesFromDb(tenant string) ([]string, error) {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	rows, err := dbConn.Model(&AliasModel{}).
		Where("cluster = ?", tenant).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aliases := make([]string, 0)
	for rows.Next() {
		var record AliasModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		if len(record.Alias) == 0 {
			break
		}

		aliases = append(aliases, record.Alias)
	}

	return aliases, nil
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

func (self *clusterImpl) updateTenantToDb(tenant kubernetes.Tenant) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

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
	aliasRecords := make([]AliasModel, 0)

	for _, alias := range tenant.GetAliases() {
		aliasRecords = append(aliasRecords, AliasModel{
			Alias:   alias,
			Cluster: tenant.GetName(),
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(aliasRecords, batchSize)
	if resp.Error != nil {
		return resp.Error
	}

	resp = dbConn.Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&ClusterModel{
			Name:       tenant.GetName(),
			Provider:   ProviderEnum(provider),
			Metadata:   string(encodedMetadata),
			Kubeconfig: string(encodedKubeconfig),
			Expire:     tenant.GetExpiredTime(),
		})
	if resp.Error == nil {
		self.tenants[tenant.GetName()] = tenant
	}

	return resp.Error
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
		if record.Expire > 0 && record.Expire < time.Now().Unix() {
			tenant, err = record.Provider.ConvertMetadataToTenant(
				record.Metadata,
				self.timeout,
			)
			if err != nil {
				return err
			}

			return self.updateTenantToDb(tenant)
		} else {
			self.tenants[record.Name] = tenant
		}
	}

	return nil
}

func (self *clusterImpl) convertAliasToTenant(
	alias string,
) (kubernetes.Tenant, error) {
	var record AliasModel

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	resp := dbConn.First(&record, "alias = ?", alias)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return self.pickTenantOrLoadFromDb(record.Cluster)
}

func (self *clusterImpl) pickTenantOrLoadFromDb(
	name string,
) (kubernetes.Tenant, error) {
	tenant, ok := self.tenants[name]
	if !ok {
		if err := self.loadTenantFromDb(name); err != nil {
			return nil, err
		}

		if tenant, ok = self.tenants[name]; !ok {
			return nil, fmt.Errorf("Can't find %s", name)
		}
	}

	return tenant, nil
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

func List(module container.Module) ([]string, error) {
	clusterMgr, ok := module.(*clusterImpl)
	if !ok {
		return nil, errors.New("Unknown module")
	}

	mapping := make(map[string]string)

	tenants, err := clusterMgr.getListTenantFromDb()
	if err != nil {
		return nil, err
	}

	for _, tenant := range tenants {
		aliases, err := clusterMgr.getListAliasesFromDb(tenant)
		if err != nil {
			return err
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
	return err
}
