package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/gorm"
	"time"
)

func InitTables(db *gorm.DB) error {
	// 严格来说，这个不是优秀实践
	return db.AutoMigrate(&User{}, &Article{}, &PublishedArticle{}, &Interactive{})
}

func InitCollection(mdb *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 创建索引
	_, err := mdb.Collection("articles").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"author_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		panic(err)
	}

	_, err = mdb.Collection("published_articles").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"author_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		panic(err)
	}
}
