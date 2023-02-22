package config

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DatabaseConfig = new(Database)
var DB *gorm.DB

type Database struct {
	Driver string
	Source string
}

func InitDatabase(cfg *viper.Viper) *Database {
	db := &Database{
		Driver: cfg.GetString("driver"),
		Source: cfg.GetString("source"),
	}
	return db
}

func ConnectMysql(driver, source string) *gorm.DB {
	if driver == "mysql" {
		db, err := gorm.Open(mysql.Open(source), &gorm.Config{})
		if err != nil {
			panic("连接数据库失败, error=" + err.Error())
		}
		return db
	} else {
		panic("数据库类型有误！")
	}

}
