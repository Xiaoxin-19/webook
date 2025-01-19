package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"
	"webok/config"
	"webok/internal/repository"
	"webok/internal/repository/dao"
	"webok/internal/service"
	"webok/internal/web"
	"webok/internal/web/middleware"
)

func main() {
	db := initDB()
	server := initServer()
	initUserHdl(db, server)
	err := server.Run(":8081")
	if err != nil {
		panic("start server failed")
	}
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

	// 设置跨域请求
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
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

	//设置cookie
	store := cookie.NewStore([]byte("123456789012345678901234567890"))
	server.Use(sessions.Sessions("ssid", store))

	//设置登录验证中间件
	loginMidl := middleware.LoginMiddlewareBuilder{}
	server.Use(loginMidl.CheckLogin())

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})

	fmt.Println(redisClient)
	return server
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	uh := web.NewUserHandler(us)
	uh.RegisterRoutes(server)
}
