package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	if server == nil {
		panic("start server failed")
	}
	useCors(server)
	useJWT(server)
	return server
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	uh := web.NewUserHandler(us)
	uh.RegisterRoutes(server)
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
