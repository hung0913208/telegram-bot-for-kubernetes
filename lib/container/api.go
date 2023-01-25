package container

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type Module interface {
	Init() error
	Deinit() error
}

type RpcModule interface {
	Module
	PairWith(module string) error
}

type wrapImpl struct {
	name   string
	module Module
	index  int
	status bool
}

type containerImpl struct {
	mapping map[string]wrapImpl
	modules []Module
}

var iContainerManager *containerImpl

func Init() error {
	if iContainerManager != nil {
		return errors.New("Only call init container one time at the begining")
	}

	iContainerManager = &containerImpl{}
	return nil
}

func Register(name string, module Module) error {
	if iContainerManager == nil {
		if err := Init(); err != nil {
			return err
		}
	}

	if iContainerManager == nil {
		return errors.New("Con't setup container manager")
	}

	if _, ok := iContainerManager.mapping[name]; ok {
		return fmt.Errorf("Object %s has been registered", name)
	}

	if err := module.Init(); err != nil {
		return err
	}

	iContainerManager.mapping[name] = wrapImpl{
		name:   name,
		module: module,
		index:  len(iContainerManager.modules),
		status: true,
	}
	iContainerManager.modules = append(iContainerManager.modules, module)
	return nil
}

func Terminate(msg string, exitCode int) {
	if iContainerManager != nil {
		for _, wrap := range iContainerManager.mapping {
			if !wrap.status {
				continue
			}

			if err := wrap.module.Deinit(); err != nil {
				log.Fatalf("%v", err)
			}

			wrap.status = false
		}
	}

	os.Exit(exitCode)
}
