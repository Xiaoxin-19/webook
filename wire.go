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
	"webok/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitDB, ioc.InitRedis,
		// DAO
		dao.NewGormUserDAO,
		// CACHE
		cache.NewCodeRedisCache, cache.NewUserCache,
		// REPO
		repository.NewCachedUserRepository, repository.NewCodeRepository,
		// Service
		ioc.InitSMSService, service.NewNormalUserService, service.NewCodeService,
		// Handler
		web.NewUserHandler,
		ioc.InitGinMiddlewares, ioc.InitWebServer,
	)
	return gin.Default()
}
