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
var userSvcProvider = wire.NewSet(
	dao.NewGormUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewNormalUserService)

var articleSvcProvider = wire.NewSet(
	repository.NewCachedArticleRepository,
	cache.NewArticleRedisCache,
	dao.NewArticleGORMDAO,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(
	dao.NewInteractiveGORMDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		thirdPartySet,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcSet,
		// CACHE
		cache.NewCodeRedisCache,
		// REPO
		repository.NewCodeRepository,
		// Service
		ioc.InitSMSService,
		service.NewCodeService,
		ioc.InitWechatService,
		// Handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisHandler,
		web.NewArticleHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		interactiveSvcSet,
		repository.NewCachedArticleRepository,
		cache.NewArticleRedisCache,
		service.NewArticleService,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service.NewInteractiveService(nil)
}
