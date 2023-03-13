package kubernetes

import (
	"errors"
)

type Pool interface {
}

type Tenant interface {
	GetName() string
	GetAliases() []string
	GetClient() (Kubernetes, error)
	GetMetadata() (interface{}, error)
	GetProvider() (string, error)
	GetKubeconfig() ([]byte, error)
	GetPool(name string) (Pool, error)
	GetExpiredTime() int64

	SetPool(name string, pool Pool) error
}

type defaultTenantImpl struct {
	name       string
	kubeconfig []byte
	metadata   interface{}
	client     Kubernetes
}

func NewDefaultTenant(
	name string,
	kubeconfig []byte,
	metadata ...interface{},
) (Tenant, error) {
	client, err := NewFromKubeconfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	if !client.Ping() {
		return nil, errors.New("can't access cluster")
	}

	if len(metadata) > 0 {
		return &defaultTenantImpl{
			name:       name,
			client:     client,
			metadata:   metadata[0],
			kubeconfig: kubeconfig,
		}, nil
	} else {
		return &defaultTenantImpl{
			name:       name,
			client:     client,
			kubeconfig: kubeconfig,
		}, nil
	}
}

func (self *defaultTenantImpl) GetName() string {
	return self.name
}

func (self *defaultTenantImpl) GetAliases() []string {
	return make([]string, 0)
}

func (self *defaultTenantImpl) GetProvider() (string, error) {
	return "", errors.New("Can't get provider from unknown tenant")
}

func (self *defaultTenantImpl) GetClient() (Kubernetes, error) {
	return self.client, nil
}

func (self *defaultTenantImpl) GetKubeconfig() ([]byte, error) {
	if _, err := self.client.GetPods(""); err != nil {
		return nil, err
	}
	return self.kubeconfig, nil
}

func (self *defaultTenantImpl) GetMetadata() (interface{}, error) {
	return nil, errors.New("Can't get metadata from unknown tenant")
}

func (self *defaultTenantImpl) GetPool(name string) (Pool, error) {
	return nil, errors.New("Can't get pool from unknown tenant")
}

func (self *defaultTenantImpl) GetExpiredTime() int64 {
	return 0
}

func (self *defaultTenantImpl) SetPool(name string, pool Pool) error {
	return errors.New("Can't set pool from unknown tenant")
}
