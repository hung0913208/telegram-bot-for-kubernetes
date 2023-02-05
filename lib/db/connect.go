package db

import (
    "context"
	"fmt"
	"time"

	sentry "github.com/getsentry/sentry-go"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Db interface {
	Establish() *gorm.DB
}

type dbImpl struct {
	dbConn *gorm.DB
    dialector gorm.Dialector

	host, username, password string
	port                     int
}

type dialectorProxyImpl struct {
    dialector gorm.Dialector
}

var (
    _ gorm.Dialector = &dialectorProxyImpl{}
)

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
            "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&" +
            "loc=Local&" +
            "interpolateParams=true",
			username, password,
			host, port,
			database,
		),
	)

	dbConn, err := gorm.Open(
		&dialectorProxyImpl{
            dialector: mysqlConn,
        },
		&gorm.Config{
			// @NOTE: most of the case we don't need transaction since we will
			//        handle everything manually so don't need to
			SkipDefaultTransaction: true,

            // @NOTE: cache prepared statement so the future queries might be
            //        speed up
            PrepareStmt: true,

			// @NOTE: configure logger
			Logger: newLogger,
		},
	)

	if err != nil {
		return nil, err
	}

	return &dbImpl{
		dbConn:    dbConn,
        dialector: mysqlConn,
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
		&dialectorProxyImpl{
            dialector: pgConn,
        },
		&gorm.Config{
			// @NOTE: most of the case we don't need transaction since we will
			//        handle everything manually so don't need to
			SkipDefaultTransaction: true,

            // @NOTE: cache prepared statement so the future queries might be
            //        speed up
            PrepareStmt: true,

			// @NOTE: configure logger
			Logger: newLogger,
		},
	)

	if err != nil {
		return nil, err
	}

	return &dbImpl{
		dbConn:    dbConn,
        dialector: pgConn,
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

func (self *dialectorProxyImpl) Name() string {
    return self.dialector.Name()
}

func (self *dialectorProxyImpl) Initialize(dbSql *gorm.DB) error {
	span := sentry.StartSpan(
		context.Background(),
		self.dialector.Name(),
		sentry.TransactionName("inittialize"),
	)
	defer span.Finish()

    return self.dialector.Initialize(dbSql)
}

func (self *dialectorProxyImpl) Migrator(dbSql *gorm.DB) gorm.Migrator {
	span := sentry.StartSpan(
		context.Background(),
		self.dialector.Name(),
		sentry.TransactionName("migrator"),
	)
	defer span.Finish()

    return self.dialector.Migrator(dbSql)
}

func (self *dialectorProxyImpl) DataTypeOf(field *schema.Field) string {
    return self.dialector.DataTypeOf(field)
}

func (self *dialectorProxyImpl) DefaultValueOf(
    field *schema.Field,
) clause.Expression {
    return self.dialector.DefaultValueOf(field)
}

func (self *dialectorProxyImpl) BindVarTo(
    writer clause.Writer,
    stmt *gorm.Statement,
    v interface{},
) {
    self.dialector.BindVarTo(writer, stmt, v)
}

func (self *dialectorProxyImpl) QuoteTo(writer clause.Writer, data string) {
    self.dialector.QuoteTo(writer, data)
}

func (self *dialectorProxyImpl) Explain(sql string, vars ...interface{}) string {
	span := sentry.StartSpan(
		context.Background(),
		self.dialector.Name(),
		sentry.TransactionName("explain"),
	)
	defer span.Finish()

    return self.dialector.Explain(sql, vars...)
}
