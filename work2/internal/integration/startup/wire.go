//go:build wireinject

package startup

import (
	"example/wb/internal/repository"
	"example/wb/internal/repository/cache"
	"example/wb/internal/repository/dao"
	"example/wb/internal/service"
	"example/wb/internal/web"
	"example/wb/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 初始化第三方依赖
		// ioc.InitFreeCache,
		InitRedis, ioc.InitDB,
		dao.NewUserDao,
		// cache部分
		cache.NewUserCache, cache.NewCodeCache,
		// repository部分
		repository.NewCachedCodeRepository, repository.NewCachedUserRepository,

		ioc.InitSMSService, service.NewCodeService, service.NewUserService,

		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()

}
