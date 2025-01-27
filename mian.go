package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"
	"webok/config"
	"webok/internal/repository"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
	"webok/internal/service"
	"webok/internal/service/sms/localsms"
	"webok/internal/web"
	"webok/internal/web/middleware"
	localMemCache "webok/pkg"
	"webok/pkg/ginx/middleware/ratelimit"
)

func main() {
	db := initDB()
	server := initServer()
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	//codeSvc := initCodeSvc(redisClient)
	localCache := localMemCache.NewLocalMemCache(10 * time.Minute)
	codeSvc := initCodeSvc(localCache)
	initUserHdl(db, redisClient, codeSvc, server)

	err := server.Run(":8081")
	if err != nil {
		panic("start server failed")
	}
}

//	func initCodeSvc(redisClient redis.Cmdable) *service.CodeService {
//		cc := cache.NewCodeRedisCache(redisClient)
//		cacheRepo := repository.NewCodeRepository(cc)
//		return service.NewCodeService(cacheRepo, localsms.NewLocalSmsService())
//	}
func initCodeSvc(memCache *localMemCache.LocalMemCache) *service.CodeService {
	cc := cache.NewCodeLocalMemCache(memCache)
	cacheRepo := repository.NewCodeRepository(cc)
	return service.NewCodeService(cacheRepo, localsms.NewLocalSmsService())
}
func initDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.Config.DB.DSN), &gorm.Config{})
	if err != nil {
		panic("data init failed")
	}
	if err = dao.InitTables(db); err != nil {
		panic("database tables init failed")
	}
	return db
}

func initServer() *gin.Engine {
	server := gin.Default()
	if server == nil {
		panic("start server failed")
	}
	useCors(server)
	useJWT(server)
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	////useRateLimit(server, redisClient)
	return server
}

func initUserHdl(db *gorm.DB, redisClient redis.Cmdable,
	codeSvc *service.CodeService, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(redisClient)
	ur := repository.NewUserRepository(ud, uc)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us, codeSvc)
	hdl.RegisterRoutes(server)
}

func useJWT(server *gin.Engine) {
	login := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

func useCors(server *gin.Engine) {
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
	server.Use(cors.New(corsConfig))
}

func useRateLimit(server *gin.Engine, redisClient *redis.Client) {
	server.Use(ratelimit.NewBuilder(redisClient,
		time.Second*10, 1).Build())
}
