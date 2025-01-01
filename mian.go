package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
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
		return
	}
}

func initDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=webook port=15432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
	corsConfig.AllowHeaders = []string{"Content-Length"}
	corsConfig.AllowOrigins = []string{"http://localhost", "http://127.0.0.1"}
	corsConfig.MaxAge = 12 * time.Hour
	server.Use(cors.New(corsConfig))

	//设置cookie
	store := cookie.NewStore([]byte("123456789012345678901234567890"))
	server.Use(sessions.Sessions("ssid", store))

	//设置登录验证中间件
	loginMidl := middleware.LoginMiddlewareBuilder{}
	server.Use(loginMidl.CheckLogin())

	return server
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	uh := web.NewUserHandler(us)
	uh.RegisterRoutes(server)
}
