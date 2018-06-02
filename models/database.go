package models

import (
	"github.com/jinzhu/gorm"
	"log"
	"../param"
)

var databaseConn *gorm.DB = nil

func Database() *gorm.DB {
	//open a db connection
	if databaseConn == nil {
		d, err := gorm.Open("mysql", param.Conf.DbUrl)
		d.LogMode(false)
		if err != nil {
			log.Println(err)
			return nil
		}
		// skip save associations of gorm -> manual save by code
		databaseConn = d.Set("gorm:save_associations", false)
		databaseConn.DB().SetMaxOpenConns(20)
		databaseConn.DB().SetMaxIdleConns(10)
	}
	return databaseConn
}
