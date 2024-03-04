package ioc

import (
	"example/wb/internal/repository/dao"
	"example/wb/pkg/logger"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
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

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(goormLoggerFunc(l.Debug), glogger.Config{
			//慢查询日志，设置为0，所有的语句都会打印出来
			SlowThreshold: 0,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic("数据库错误")
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type goormLoggerFunc func(msg string, field ...logger.Field)

func (g goormLoggerFunc) Printf(msg string, field ...interface{}) {
	g(msg, logger.Field{
		Key: "args",
		Val: field,
	})
}
