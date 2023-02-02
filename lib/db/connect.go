package db

import (
    "fmt"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/driver/mysql"
    "gorm.io/gorm/logger"
    "gorm.io/gorm"
)

type Db interface {
    Establish() *gorm.DB
}

type dbImpl struct {
    dbConn *gorm.DB

    host, username, password string
    port int
}

func NewMysql(
    host string, 
    port int,
    username, password, database string,
    timeout time.Duration,
) (Db, error) {
    newLogger := newGormSentryLogger(
        logger.Config{
            SlowThreshold:             timeout,
            LogLevel:                  logger.Silent, 
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    )

    mysqlConn := mysql.Open(
        fmt.Sprintf(
            "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            username, password,
            host, port,
            database,
        ),
    )
 
    dbConn, err := gorm.Open(
        mysqlConn, 
        &gorm.Config{
            // @NOTE: most of the case we don't need transaction since we will 
            //        handle everything manually so don't need to 
            SkipDefaultTransaction: true,

            // @NOTE: configure logger 
            Logger: newLogger,
        },
    )

    if err != nil {
        return nil, err
    }

    return &dbImpl{
        dbConn: dbConn,
    }, nil
}

func NewPg(
    host string, 
    port int,
    username, password, database string,
    timeout time.Duration,
) (Db, error) {
    newLogger := newGormSentryLogger(
        logger.Config{
            SlowThreshold:             timeout,
            LogLevel:                  logger.Silent, 
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    )

    pgConn := postgres.Open(
        fmt.Sprintf(
            "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
            host, 
            username, password, 
            database, 
            port,
        ),
    )

    dbConn, err := gorm.Open(
        pgConn,
        &gorm.Config{
            // @NOTE: most of the case we don't need transaction since we will 
            //        handle everything manually so don't need to 
            SkipDefaultTransaction: true,

            // @NOTE: configure logger 
            Logger: newLogger,
        },
    )

    if err != nil {
        return nil, err
    }

    return &dbImpl{
        dbConn: dbConn,
    }, nil
}

func (self *dbImpl) setupConnectionPool(
    maxIdleConnection, maxOpenConnection int,
    maxLifetime time.Duration,
) error {
    sqlDB, err := self.dbConn.DB()
    if err != nil {
        return err
    }
    
    sqlDB.SetMaxIdleConns(maxIdleConnection) 
    sqlDB.SetMaxOpenConns(maxOpenConnection) 
    sqlDB.SetConnMaxLifetime(maxLifetime)
    return nil
}

func (self *dbImpl) Establish() *gorm.DB {
    return self.dbConn
}
