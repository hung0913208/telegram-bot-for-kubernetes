package kubernetes

import (
	"errors"
)

type Pool interface {
}

type Tenant interface {
	GetName() string
	GetMedadata() (interface{}, error)
	GetProvider() (string, error)
	GetKubeconfig() (string, error)
	GetPool(name string) (Pool, error)

	SetPool(name string, pool Pool) error
}

type defaultTenantImpl struct {
	name       string
	kubeconfig string
	metadata   interface{}
	client     Kubernetes
}

func NewDefaultTenant(
	name string,
	kubeconfig string,
	metadata ...interface{},
) (Tenant, error) {
	client, err := NewFromKubeconfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	if _, err := client.GetPods(); err != nil {
		return nil, err
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

func (self *defaultTenantImpl) GetProvider() (string, error) {
	return "", errors.New("Can't get provider from unknown tenant")
}

func (self *defaultTenantImpl) GetKubeconfig() (string, error) {
	if _, err := self.client.GetPods(); err != nil {
		return "", err
	}
	return self.kubeconfig, nil
}

func (self *defaultTenantImpl) GetMedadata() (interface{}, error) {
	return nil, errors.New("Can't get metadata from unknown tenant")
}

func (self *defaultTenantImpl) GetPool(name string) (Pool, error) {
	return nil, errors.New("Can't get pool from unknown tenant")
}

func (self *defaultTenantImpl) SetPool(name string, pool Pool) error {
	return errors.New("Can't set pool from unknown tenant")
}
