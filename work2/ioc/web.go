package ioc

import (
	"example/wb/internal/web"
	login "example/wb/internal/web/middleware"
	"example/wb/pkg/ginx/middleware/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
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
		}),
		ratelimit.NewBuilder(redisClient, time.Second, 1000).Build(),
		(&login.LoginMiddlewareBuilder{}).CheckJWTLogin(),
	}

}
