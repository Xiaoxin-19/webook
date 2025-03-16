//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webok/internal/events/article"
	"webok/internal/repository"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
	"webok/internal/service"
	"webok/internal/web"
	ijwt "webok/internal/web/jwt"
	"webok/ioc"
)

func InitWebServer() *App {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,

		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		// DAO
		dao.NewGormUserDAO, dao.NewArticleGORMDAO, dao.NewInteractiveGORMDAO,
		// CACHE
		cache.NewCodeRedisCache, cache.NewUserCache, cache.NewArticleRedisCache,
		cache.NewRedisInteractiveCache,
		// REPO
		repository.NewCachedUserRepository, repository.NewCodeRepository, repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,
		// Service
		ioc.InitSMSService, service.NewNormalUserService, service.NewCodeService,
		ioc.InitWechatService, service.NewArticleService, service.NewInteractiveService,
		// Handler
		ijwt.NewRedisHandler, web.NewUserHandler, web.NewOAuth2WechatHandler, web.NewArticleHandler,
		ioc.InitGinMiddlewares, ioc.InitWebServer,
		wire.Struct(new(App), "*"),
	)

	return new(App)
}
