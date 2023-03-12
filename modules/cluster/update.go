package cluster

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm/clause"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/platform"
)

func (self *clusterImpl) updateDeploymentToDb(tenant kubernetes.Tenant) error {
	return nil
}

func (self *clusterImpl) updatePodToDb(tenant kubernetes.Tenant) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
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

	records := make([]PodModel, 0)

	for _, pod := range pods.Items {
		records = append(records, PodModel{
			BaseModel: platform.BaseModel{
				UUID: pod.Name,
			},
			Deployment:    "",
			Status:        1,
			Version:       "",
			CpuLimit:      1,
			MemoryLimit:   1,
			CpuRequest:    1,
			MemoryRequest: 1,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(records, batchSize)
	return resp.Error
}

func (self *clusterImpl) updateVolumeToDb(tenant kubernetes.Tenant) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	client, err := tenant.GetClient()
	if err != nil {
		return err
	}

	pvs, err := client.GetPVs()
	if err != nil {
		return err
	}

	records := make([]VolumeModel, 0)

	for _, pv := range pvs.Items {
		records = append(records, VolumeModel{
			BaseModel: platform.BaseModel{
				UUID: fmt.Sprintf("%s-%s", tenant.GetName(), pv.Name),
			},
			Pod:      "",
			Name:     pv.Name,
			Cluster:  tenant.GetName(),
			Usage:    1,
			Capacity: 1,
		})
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		Clauses(clause.OnConflict{UpdateAll: true}).
		CreateInBatches(records, batchSize)
	return resp.Error
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
