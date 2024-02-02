package main

import (
	"example/wb/config"
	"example/wb/internal/repository"
	"example/wb/internal/repository/dao"
	"example/wb/internal/service"
	"example/wb/internal/web"
	login "example/wb/internal/web/middleware"
	"example/wb/pkg/ginx/middleware/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := InitDB()

	server := InitWebServer(db)
	// server := gin.Default()
	// server.GET("/hello", func(ctx *gin.Context) {
	// // 	ctx.String(http.StatusOK, "Hello")
	// })

	InitUserHandler(db, server)

	server.Run(":8081")
}

func InitDB() *gorm.DB {
	// dsn := "root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
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
		// "Authorization" 前端约定头部
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true, // cookies一些东西允许带过来
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(origin, "localhost")
		},
		MaxAge: 12 * time.Hour,
	}))
	login := &login.LoginMiddlewareBuilder{}
	// cookie存储session
	// store := cookie.NewStore([]byte("secret"))
	// memstore单机单实例存储
	// store := memstore.NewStore([]byte("mY2gT5iP0xZ9eX7tZ5eU9zI4lW0xP0wI"),
	// 	[]byte("kL2jF9cF3sL7pX3zZ6qX6vX4kL7qS2xQ"))
	// redis存储session
	// store, err := redis.NewStore(16, "tcp",
	// 	"localhost:16379", "",
	// 	[]byte("mY2gT5iP0xZ9eX7tZ5eU9zI4lW0xP0wI"),
	// 	[]byte("kL2jF9cF3sL7pX3zZ6qX6vX4kL7qS2xQ"))

	// if err != nil {
	// 	panic(err)
	// }

	// store := memstore.NewStore([]byte("mY2gT5iP0xZ9eX7tZ5eU9zI4lW0xP0wI"), []byte("kL2jF9cF3sL7pX3zZ6qX6vX4kL7qS2xQ"))
	// redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:16379",
	// 	Password: "",
	// })
	redisClient := redis.NewClient(&redis.Options{Addr: config.Config.Redis.Addr})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	server.Use(login.CheckJWTLogin())

	// server.Use(sessions.Sessions("ssid", store), login.CheckJWTLogin())
	// server.Use(, login.CheckJWTLogin())
	return server
}

func InitUserHandler(db *gorm.DB, server *gin.Engine) {
	udao := dao.NewUserDao(db)
	urepo := repository.NewUserRepository(udao)
	usvc := service.NewUserService(urepo)
	hdl := web.NewUserHandler(usvc)
	hdl.RegisterRoutes(server)
}
