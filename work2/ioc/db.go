package ioc

import (
	"example/wb/config"
	"example/wb/internal/repository/dao"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
	if err != nil {
		panic("数据库错误")
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
