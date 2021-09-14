package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var option = MySQLOption{
	Host:     "localhost",
	Port:     3306,
	User:     "",
	Password: "",
	Database: "",
}

func TestCtxDB(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()
	db, _ := gorm.Open(mysql.Open(option.DSN()), &gorm.Config{})

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db = db.Table("id_gens").Where("object_type IN ?", []string{"COMMAND", "INVOCATION"})
	for i := 0; i < 100; i++ {
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			db = db.WithContext(ctx)
			result := map[string]interface{}{}
			ret := db.Where("max_id = ?", index).Take(&result)
			if ret.Error == nil {
				fmt.Println(result)
			} else {
				fmt.Println(ret.Error)
			}
		}()
	}
	wg.Wait()
}

func TestConnectionPool(t *testing.T) {
	var wg sync.WaitGroup
	db, _ := gorm.Open(mysql.Open(option.DSN()), &gorm.Config{})

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(1)
	// sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db = db.Exec("select sleep(10);")
			if db.Error == nil {
				fmt.Println("sleep finish")
			} else {
				fmt.Println(db.Error)
			}
		}()
	}
	wg.Wait()
}

func TestConnectionTimeout(t *testing.T) {
	var timeoutOption = MySQLOption{
		Host:           "127.0.0.2",
		Port:           3306,
		User:           "",
		Password:       "",
		Database:       "",
		ConnectTimeout: 10,
	}
	fmt.Println("start connect-" + time.Now().String())
	db, _ := gorm.Open(mysql.Open(timeoutOption.DSN()), &gorm.Config{})
	fmt.Println("connect end-" + time.Now().String())
	if db != nil {
		t.Error("build connection success?")
	}
}
