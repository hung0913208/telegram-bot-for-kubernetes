package db

import (
	"errors"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"

	"gorm.io/gorm"
)

type dbModuleImpl struct {
	dbObj Db

	initCallback func(module *dbModuleImpl, timeout time.Duration) error
}

var (
	_ container.Module = &dbModuleImpl{}
)

func (self *dbModuleImpl) Init(timeout time.Duration) error {
	return self.initCallback(self, timeout)
}

func (self *dbModuleImpl) Deinit() error {
	return nil
}

func (self *dbModuleImpl) Execute(args []string) error {
	return errors.New("This abstract module doesn't support this function")
}

func NewMysqlModule(
	host string,
	port int,
	username, password, database string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewMysql(host, port, username, password, database, timeout)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}

func NewPgModule(
	host string,
	port int,
	username, password, database string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewPg(host, port, username, password, database, timeout)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}

func Establish(module container.Module) (*gorm.DB, error) {
	if wrapper, ok := module.(*dbModuleImpl); ok {
		return wrapper.dbObj.Establish(), nil
	}
	return nil, errors.New("Can't cast module to database module")
}
