package startup

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
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

func InitMongoDB() *mongo.Database {
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-1",
		AuthSource:    "admin",
		Username:      "root",
		Password:      "example",
	}
	defer cancel()
	opt := options.Client().ApplyURI("mongodb://localhost:27017").SetMonitor(monitor).SetAuth(credential)
	client, err := mongo.Connect(opt)
	if err != nil {
		panic(err)
	}
	// 验证连接
	if err := client.Ping(ctx, nil); err != nil {
		panic(err)
	}
	return client.Database("webook")
}
