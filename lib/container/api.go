package container

import (
	"errors"
	"fmt"
	"log"
	"os"
    "time"
)

type Module interface {
	Init(timeout time.Duration) error
	Deinit() error
    Execute(args []string) error
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

	iContainerManager = &containerImpl{
		mapping: make(map[string]wrapImpl),
		modules: make([]Module, 0),
	}
	return nil
}

func RegisterSimpleModule(name string, module Module, 
                          timeout int) error {
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

	if err := module.Init(time.Duration(timeout)*time.Millisecond); err != nil {
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

func RegisterRpcModule(name string, module Module,
                       timeout int) error {
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

func Pick(name string) (Module, error) {
    wrapper, ok := iContainerManager.mapping[name]

    if !ok {
        return nil, fmt.Errorf("Module `%s` doesn`t exist", name)
    }
    return wrapper.module, nil
}

func Lookup(index int) (Module, error) {
    if index >= len(iContainerManager.modules) {
        return nil, fmt.Errorf(
            "index `%d` is out of scope, must below %d", 
            index, 
            len(iContainerManager.modules),
        )
    }

    return iContainerManager.modules[index], nil
}

