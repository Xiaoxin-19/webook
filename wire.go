//go:build wireinject

package main

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

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		// DAO
		dao.NewGormUserDAO, dao.NewArticleGORMDAO,
		// CACHE
		cache.NewCodeRedisCache, cache.NewUserCache, cache.NewArticleRedisCache,
		// REPO
		repository.NewCachedUserRepository, repository.NewCodeRepository, repository.NewCachedArticleRepository,
		// Service
		ioc.InitSMSService, service.NewNormalUserService, service.NewCodeService,
		ioc.InitWechatService, service.NewArticleService,
		// Handler
		ijwt.NewRedisHandler, web.NewUserHandler, web.NewOAuth2WechatHandler, web.NewArticleHandler,
		ioc.InitGinMiddlewares, ioc.InitWebServer,
	)
	return gin.Default()
}
