package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"songof.com/edge-storage-core/storage"
	"time"
)

const (
	DefaultMaxOpenConns    = 100
	DefaultConnMaxLifetime = time.Hour
)

// var db *gorm.DB
var db *Database

type Condition = func(db *gorm.DB) *gorm.DB

type MySQLOption struct {
	Host            string
	Port            uint
	User            string
	Password        string
	Database        string
	ConnectTimeout  int
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Database struct {
	Option *MySQLOption
	gormDB *gorm.DB
	rawDB  *sql.DB
}

func (option *MySQLOption) DSN() (dsn string) {
	var timeout int
	if option.ConnectTimeout > 0 {
		timeout = option.ConnectTimeout
	} else {
		timeout = 10
	}
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds",
		option.User, option.Password, option.Host, option.Port, option.Database, timeout)
	return
}

func Init(option MySQLOption, p logger.Interface) {
	if db != nil {
		return
	}
	gormDB, err := gorm.Open(mysql.Open(option.DSN()), &gorm.Config{
		Logger: p,
	})
	if err != nil {
		storage.DatabaseErrorInc()
		fmt.Println("gorm open failed!")
		panic(err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		storage.DatabaseErrorInc()
		panic(err)
	}

	//强制设置最大连接数, 避免极端情况mysql连接用尽
	maxOpenConns := DefaultMaxOpenConns
	if option.MaxOpenConns > 0 {
		maxOpenConns = option.MaxOpenConns
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(option.MaxIdleConns)

	connMaxLifetime := DefaultConnMaxLifetime
	if option.ConnMaxLifetime > 0 {
		connMaxLifetime = option.ConnMaxLifetime
		// 设置最小下限为1min, 避免时间太短没有意义
		if connMaxLifetime < time.Minute {
			connMaxLifetime = time.Minute
		}
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	db = &Database{
		Option: &option,
		gormDB: gormDB,
		rawDB:  nil,
	}
}

func DB() *gorm.DB {
	return db.gormDB
}

func CtxDB(ctx context.Context) *gorm.DB {
	return db.gormDB.WithContext(ctx)
}

func Raw() *sql.DB {
	if db.rawDB == nil {
		db.rawDB, _ = sql.Open("mysql", db.Option.DSN())
	}
	return db.rawDB
}
