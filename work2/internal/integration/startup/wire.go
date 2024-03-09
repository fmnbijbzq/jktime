//go:build wireinject

package startup

import (
	"example/wb/internal/repository"
	"example/wb/internal/repository/cache"
	"example/wb/internal/repository/dao"
	"example/wb/internal/service"
	"example/wb/internal/web"
	ijwt "example/wb/internal/web/jwt"
	"example/wb/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 初始化第三方依赖
		ioc.InitFreeCache,
		ioc.InitRedis, ioc.InitDB,
		ioc.InitWechatService,
		ioc.InitLogger,

		dao.NewUserDao, dao.NewSmsDao,
		// cache部分
		cache.NewUserCache, cache.NewCodeLocalCache,
		// repository部分
		repository.NewCachedCodeRepository, repository.NewCachedUserRepository,
		repository.NewAsyncSMSRepository,
		// service部分
		ioc.InitSMSService, service.NewCodeService, service.NewUserService,
		// web部分
		web.NewUserHandler, web.NewOAuth2WechatHandler, ijwt.NewJwtHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()

}

//	func InitUserHandler(dao dao.ArticleDao) *web.UserHandler {
//		wire.Build()
//		return &web.UserHandler{}
//	}
func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		ioc.InitLogger,
		repository.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

// func InitArticleHandler(dao dao.ArticleDao) *web.ArticleHandler {
// 	wire.Build(

// 		ioc.InitLogger,

// 		service.NewArticleService,

// 		repository.NewArticleRepository,
// 	)
// 	return &web.ArticleHandler{}

// }
