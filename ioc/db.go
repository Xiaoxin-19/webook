package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"webok/internal/repository/dao"
	"webok/pkg/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	url := viper.GetString("database.Url")

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: glogger.New(gormLogger(l.Debug), glogger.Config{
			// 慢查询
			SlowThreshold: 0,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic("data init failed")
	}
	if err = dao.InitTables(db); err != nil {
		panic("database tables init failed")
	}
	return db
}

type gormLogger func(msg string, fields ...logger.Field)

func (g gormLogger) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Val: args})
}
