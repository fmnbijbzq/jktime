package ioc

import (
	"context"
	"example/wb/internal/web"
	"example/wb/internal/web/jwt"
	"example/wb/internal/web/middleware"
	"example/wb/pkg/ginx/middleware/ratelimit"
	"example/wb/pkg/limiter"
	"example/wb/pkg/logger"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable,
	hdl jwt.Handler, l logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			// AllowOrigins:     []string{"https://localhost:3000"},
			// AllowMethods:     []string{"POST"}, // 不配置默认全部
			// "Authorization" 前端约定头部
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
			AllowCredentials: true, // cookies一些东西允许带过来
			AllowOriginFunc: func(origin string) bool {
				return strings.Contains(origin, "localhost")
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(limiter.NewRedisLimter(redisClient, time.Second, 1000)).Build(),
		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("这是在debug", logger.Field{Key: "req", Val: al})
		}).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginMiddlewareBuilder(hdl).CheckJWTLogin(),
	}

}
