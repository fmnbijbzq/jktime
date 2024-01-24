package main

import (
	"example/wb/internal/repository"
	"example/wb/internal/repository/dao"
	"example/wb/internal/service"
	"example/wb/internal/web"
	login "example/wb/internal/web/middleware"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := InitDB()

	server := InitWebServer(db)
	InitUserHandler(db, server)

	server.Run(":8080")
}

func InitDB() *gorm.DB {
	dsn := "root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库错误")
	}
	dao.InitTables(db)

	return db
}

func InitWebServer(db *gorm.DB) *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://localhost:3000"},
		// AllowMethods:     []string{"POST"}, // 不配置默认全部
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true, // cookies一些东西允许带过来
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(origin, "localhost")
		},
		MaxAge: 12 * time.Hour,
	}))
	login := &login.LoginMiddlewareBuilder{}
	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
	return server
}

func InitUserHandler(db *gorm.DB, server *gin.Engine) {
	udao := dao.NewUserDao(db)
	urepo := repository.NewUserRepository(udao)
	usvc := service.NewUserService(urepo)
	hdl := web.NewUserHandler(usvc)
	hdl.RegisterRoutes(server)
}
