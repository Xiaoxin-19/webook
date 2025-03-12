package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webok/internal/web"
	ijwt "webok/internal/web/jwt"
	"webok/internal/web/middleware"
	"webok/pkg/ginx/middleware/ratelimit"
	"webok/pkg/limiter"
	"webok/pkg/logger"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, wechatHandler *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHandler.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(client redis.Cmdable, jwt ijwt.Handler, l logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		useCors(),
		useJWT(jwt),
		useRateLimit(client),
		useLogger(l),
		useErrorLogger(l),
	}
}

func useCors() gin.HandlerFunc {
	// 设置跨域请求
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	// 这个是允许前端访问你的后端响应中带的头部
	corsConfig.ExposeHeaders = []string{"x-jwt-token", "x-refresh-token"}
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

func useJWT(jwt ijwt.Handler) gin.HandlerFunc {
	login := middleware.LoginJWTMiddlewareBuilder{
		Handler: jwt,
	}
	return login.CheckLogin()
}

func useLogger(l logger.Logger) gin.HandlerFunc {
	LogFn := func(ctx context.Context, log *middleware.AccessLog) {
		l.Debug("", logger.Field{Val: log})
	}
	return middleware.NewLoggerMiddlewareBuilder(LogFn).
		WithReqBody().
		WithRespBody().
		Build()
}

func useErrorLogger(l logger.Logger) gin.HandlerFunc {
	return middleware.NewErrorLoggerBuilder(l).Build()
}
