package database

import (
	"fmt"
	"github.com/CoinSummer/go-base/db/mysql"
	"github.com/jinzhu/gorm"
	"log"
)

func init() {
}

func NewMysql(mysqlConfig *mysql.Config) *mysql.MySQL {
	mysqlClient, err := mysql.New(mysqlConfig)
	if err != nil {
		log.Fatalf("new mysql client error:%s", err)
	}

	url := mysqlConfig.User + ":" + mysqlConfig.Pass + "@tcp(" + mysqlConfig.Host + ")/" + mysqlConfig.Db + "?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local"
	db1, err := gorm.Open("mysql", url)
	if err != nil {
		fmt.Println("error")
	}

	db1.DB().SetMaxIdleConns(mysqlConfig.MaxIdleConn)
	db1.DB().SetMaxOpenConns(mysqlConfig.MaxOpenConn)
	db1.DB().SetConnMaxIdleTime(mysqlConfig.MaxIdleTimeConn)
	db1.DB().SetConnMaxLifetime(mysqlConfig.MaxLifeTimeConn)

	mysqlClient.Client = db1
	return mysqlClient
}
