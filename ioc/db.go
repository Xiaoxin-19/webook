package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"webok/internal/repository/dao"
)

func InitDB() *gorm.DB {
	url := viper.GetString("database.Url")
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		panic("data init failed")
	}
	if err = dao.InitTables(db); err != nil {
		panic("database tables init failed")
	}
	return db
}
