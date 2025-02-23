//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webok/internal/repository"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
	"webok/internal/service"
	"webok/internal/web"
	ijwt "webok/internal/web/jwt"
	"webok/ioc"
)

var thirdPartySet = wire.NewSet(
	InitDB, InitRedis, InitLogger,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		thirdPartySet,
		// DAO
		dao.NewGormUserDAO, dao.NewArticleGORMDAO,
		// CACHE
		cache.NewCodeRedisCache, cache.NewUserCache,
		// REPO
		repository.NewCachedUserRepository, repository.NewCodeRepository,
		repository.NewCachedArticleRepository,
		// Service
		ioc.InitSMSService, service.NewNormalUserService, service.NewCodeService,
		service.NewArticleService,
		// Handler
		web.NewUserHandler, web.NewOAuth2WechatHandler, ijwt.NewRedisHandler, web.NewArticleHandler,
		ioc.InitGinMiddlewares, ioc.InitWebServer, ioc.InitWechatService,
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		//第三方依赖
		thirdPartySet,
		// DAO
		dao.NewArticleGORMDAO,
		// REPO
		repository.NewCachedArticleRepository,
		// Service
		service.NewArticleService,
		// Handler
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
