package cluster

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
)

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

		if len(name) > 0 && name != record.Name {
			continue
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
