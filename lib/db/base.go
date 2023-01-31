package db

import (
)

type Db interface {

}

func NewPgModule() {
}

func setupConnectionPool(
    maxIdleConnection, maxOpenConnection int,
) {
//    sqlDB, err := db.DB()
//    
//    // SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
//    sqlDB.SetMaxIdleConns(10)
//    
//    // SetMaxOpenConns sets the maximum number of open connections to the database.
//    sqlDB.SetMaxOpenConns(100)
//    
//    // SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
//    sqlDB.SetConnMaxLifetime(time.Hour)
}
