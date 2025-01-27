package ioc

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"webok/config"
	"webok/internal/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.Config.DB.DSN), &gorm.Config{})
	if err != nil {
		panic("data init failed")
	}
	if err = dao.InitTables(db); err != nil {
		panic("database tables init failed")
	}
	return db
}
