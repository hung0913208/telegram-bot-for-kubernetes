package cluster

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
)

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
