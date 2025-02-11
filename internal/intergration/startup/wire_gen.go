// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"webok/internal/repository"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
	"webok/internal/service"
	"webok/internal/web"
	"webok/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	v := ioc.InitGinMiddlewares(cmdable)
	db := ioc.InitDB()
	userDAO := dao.NewGormUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewNormalUserService(userRepository)
	codeCache := cache.NewCodeRedisCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	wechatService := ioc.InitWechatService()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler)
	return engine
}
