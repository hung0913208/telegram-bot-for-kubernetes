package bizfly

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/bizflycloud/gobizfly"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
)

const SevenDaysDuration int64 = int64(7 * 24 * int(time.Hour/time.Second))

type tenantImpl struct {
	name       string
	provider   Api
	kubeconfig []byte
	client     kubernetes.Kubernetes
	cluster    *api.Cluster
}

type tenantMetadataImpl struct {
	Account string `json:"account_id"`
	Cluster string `json:"cluster_id"`
	Host    string `json:"host"`
	Region  string `json:"region"`
}

func NewTenant(
	provider Api,
	cluster *api.Cluster,
	writeable ...bool,
) (kubernetes.Tenant, error) {
	if cluster.ProvisionStatus != "PROVISIONED" {
		return nil, errors.New("Cluster must be in state `provisioned`")
	}

	kubeconfig, err := provider.GetKubeconfig(cluster.UID)
	if err != nil {
		return nil, fmt.Errorf("Get kubeconfig fails: %v", err)
	}

	client, err := kubernetes.NewFromKubeconfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf(
			"New kubernetes client fails: %v (receive %d)",
			err,
			len(kubeconfig),
		)
	}

	if err = client.SetReadable(true); err != nil {
		return nil, err
	}

	if len(writeable) > 0 {
		if err = client.SetWriteable(writeable[0]); err != nil {
			return nil, err
		}
	}

	return &tenantImpl{
		provider:   provider,
		cluster:    cluster,
		kubeconfig: []byte(kubeconfig),
		client:     client,
		name:       cluster.Name,
	}, nil
}

func NewTenantFromMetadata(
	metadata interface{},
	timeout time.Duration,
) (kubernetes.Tenant, error) {
	var account AccountModel

	properties, ok := metadata.(tenantMetadataImpl)
	if !ok {
		return nil, errors.New("this is not bizfly metadata")
	}

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	resp := dbConn.Model(AccountModel{}).
		Where("uuid = ?", properties.Account).
		First(&account)
	if resp.Error != nil {
		return nil, resp.Error
	}

	client, err := newApiWithProjectIdAndUpdateDb(
		properties.Host,
		properties.Region,
		account.ProjectId,
		account.Email,
		account.Password,
		timeout,
		false,
	)
	if err != nil {
		return nil, err
	}

	clientImpl, ok := client.(*apiImpl)
	if !ok {
		return nil, errors.New("Can't get correct bizfly API implementer")
	}
	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		time.Millisecond*timeout,
	)
	defer cancelFunc()

	cluster, err := clientImpl.client.KubernetesEngine.Get(
		ctx,
		properties.Cluster,
	)
	if err != nil {
		return nil, err
	}

	locked, err := client.IsClusterLocked(properties.Cluster)
	if err != nil {
		return nil, err
	}
	return NewTenant(client, &cluster.ExtendedCluster.Cluster, locked)
}

func (self *tenantImpl) GetClient() (kubernetes.Kubernetes, error) {
	return self.client, nil
}

func (self *tenantImpl) GetName() string {
	return self.name
}

func (self *tenantImpl) GetAliases() []string {
	return self.cluster.Tags
}

func (self *tenantImpl) GetKubeconfig() ([]byte, error) {
	return self.kubeconfig, nil
}

func (self *tenantImpl) GetExpiredTime() int64 {
	return time.Now().Unix() + SevenDaysDuration
}

func (self *tenantImpl) GetProvider() (string, error) {
	return "bizfly", nil
}

func (self *tenantImpl) GetMetadata() (interface{}, error) {
	return &tenantMetadataImpl{
		Account: self.provider.GetAccount(),
		Cluster: self.cluster.UID,
		Host:    self.provider.GetHost(),
		Region:  self.provider.GetRegion(),
	}, nil
}

func (self *tenantImpl) GetPool(name string) (kubernetes.Pool, error) {
	return nil, nil
}

func (self *tenantImpl) SetPool(name string, pool kubernetes.Pool) error {
	return nil
}
