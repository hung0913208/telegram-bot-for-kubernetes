package cluster

import (
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
)

type Cluster interface {
	container.Module
}

type clusterImpl struct {
}

func NewModule() Cluster {
	return &clusterImpl{}
}

func (self *clusterImpl) Init() error {
	// @TODO: please fill this one
	return nil
}

func (self *clusterImpl) Deinit() error {
	// @TODO: please fill this one
	return nil
}
