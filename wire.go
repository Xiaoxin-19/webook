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
		dao.NewGormUserDAO,
		// CACHE
		cache.NewCodeRedisCache, cache.NewUserCache,
		// REPO
		repository.NewCachedUserRepository, repository.NewCodeRepository,
		// Service
		ioc.InitSMSService, service.NewNormalUserService, service.NewCodeService,
		ioc.InitWechatService,
		// Handler
		ijwt.NewRedisHandler, web.NewUserHandler, web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares, ioc.InitWebServer,
	)
	return gin.Default()
}
