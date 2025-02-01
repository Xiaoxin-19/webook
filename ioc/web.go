package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webok/internal/web"
	"webok/internal/web/middleware"
	"webok/pkg/ginx/middleware/ratelimit"
	"webok/pkg/limiter"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(client redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		useCors(),
		useJWT(),
		useRateLimit(client),
	}
}

func useCors() gin.HandlerFunc {
	// 设置跨域请求
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	// 这个是允许前端访问你的后端响应中带的头部
	corsConfig.ExposeHeaders = []string{"x-jwt-token"}
	corsConfig.AllowOriginFunc = func(origin string) bool {
		if strings.HasPrefix(origin, "http://localhost") {
			return true
		}
		if strings.HasPrefix(origin, "http://127.0.0.1") {
			return true
		}
		return false
	}
	corsConfig.MaxAge = 12 * time.Hour
	return cors.New(corsConfig)
}

func useRateLimit(redisClient redis.Cmdable) gin.HandlerFunc {
	return ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build()
}

func useJWT() gin.HandlerFunc {
	login := middleware.LoginJWTMiddlewareBuilder{}
	return login.CheckLogin()
}
