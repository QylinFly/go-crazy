/**
 * 数据库配置和初始化方法
 * File: database.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 2, 2017-12-19 8:46:49 pm
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 2, 2017-12-19 8:47:05 pm
 * -----
 * Copyright 2017 - 2027 乐编程, 乐编程
 */

package Config

import (
	"os"
	"fmt"
	"path/filepath"
    "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/xoxo/crm-x/util/logger"
	//  _ "github.com/jinzhu/gorm/dialects/postgres"
	//  _ "github.com/jinzhu/gorm/dialects/sqlite"
	//  _ "github.com/jinzhu/gorm/dialects/mssql"
)

var DB	*gorm.DB


func CloseDB() {
	DB.Close()
	logger.Info("Begin to close db connection!")
}

func InitDB() {
	var err error
	if DB, err = openConnection(); err != nil {
		panic(fmt.Sprintf("\n No error should happen when connecting to test database, but got err=%+v", err))
	}

	if os.Getenv("DEBUG") == "true" {
		DB.LogMode(true)
	}
	DB.DB().SetMaxIdleConns(10)

	// runMigration()
	defer DB.Close()
}

func openConnection() (db *gorm.DB, err error) {

	var db_connection = Config.DB_CONNECTION
	var db_user = Config.DB.User
	var db_pwd = Config.DB.Password
	var db_db_name = Config.DB.Name
	var db_host = Config.DB.Host
	var db_port = Config.DB.Port

	switch db_connection {
	case "mysql":
		// db, err := gorm.Open("mysql", "root:123456@(127.0.0.1:3306)/dbname?charset=utf8&parseTime=True&loc=Local")
		var str_open = db_user+":"+db_pwd+"@("+db_host+":"+db_port+")/"+db_db_name+"?charset=utf8&parseTime=True&loc=Local"
		logger.Debug("open da str ==> "+str_open)
		db, err = gorm.Open("mysql", str_open)
	case "postgres":
		// todo
		fmt.Println("testing postgres...")
		dbhost := os.Getenv("GORM_DBHOST")
		if dbhost != "" {
			dbhost = fmt.Sprintf("host=%v ", dbhost)
		}
		db, err = gorm.Open("postgres", fmt.Sprintf("%vuser=gorm password=gorm DB.name=gorm sslmode=disable", dbhost))
	case "foundation":
		// todo
		fmt.Println("testing foundation...")
		db, err = gorm.Open("foundation", "dbname=gorm port=15432 sslmode=disable")
	case "mssql":
		// todo
		fmt.Println("testing mssql...")
		db, err = gorm.Open("mssql", "sqlserver://gorm:LoremIpsum86@localhost:1433?database=gorm")
	default:
		// todo
		fmt.Println("testing sqlite3...")
		db, err = gorm.Open("sqlite3", filepath.Join(os.TempDir(), "gorm.db"))
	}
	return
}
