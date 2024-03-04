package ioc

import (
	"example/wb/internal/repository/dao"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `json:"dsn"`
	}

	var cfg Config = Config{
		DSN: "localhost:3306",
	}
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic("数据库错误")
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
