package startup

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"webok/internal/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres password=postgres dbname=webook port=15432 sslmode=disable TimeZone=Asia/Shanghai"), &gorm.Config{})
	if err != nil {
		panic("data init failed")
	}
	if err = dao.InitTables(db); err != nil {
		panic("database tables init failed")
	}
	return db
}
