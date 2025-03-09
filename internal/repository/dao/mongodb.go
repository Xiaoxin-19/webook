package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
	"webok/internal/domain"
)

type MongoDBArticleDao struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {

	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDao) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDao) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDao) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.ID = m.node.Generate().Int64()
	article.Ctime = now
	article.Utime = now
	_, err := m.col.InsertOne(ctx, article)
	return article.ID, err
}

func (m *MongoDBArticleDao) UpdateById(ctx context.Context, entity Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{{"id", entity.ID}, {"author_id", entity.AuthorId}}
	update := bson.D{{"$set", bson.D{
		{"title", entity.Title},
		{"content", entity.Content},
		{"utime", now},
		{"status", entity.Status},
	}}}
	_, err := m.col.UpdateOne(ctx, filter, update)
	return err
}

func (m *MongoDBArticleDao) Sync(ctx context.Context, article Article) (int64, error) {
	var (
		id  = article.ID
		err error
	)

	if id > 0 {
		err = m.UpdateById(ctx, article)
	} else {
		id, err = m.Insert(ctx, article)
	}

	if err != nil {
		return 0, err
	}
	article.ID = id
	now := time.Now().UnixMilli()
	article.Utime = now

	filter := bson.D{bson.E{"id", article.ID}, bson.E{"author_id", article.AuthorId}}
	set := bson.D{
		{"$set", bson.D{
			{"title", article.Title},
			{"content", article.Content},
			{"utime", now},
			{"status", article.Status},
		}},
		{"$setOnInsert", bson.D{
			{"ctime", now},
			{"id", article.ID},
			{"author_id", article.AuthorId},
		}},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err = m.liveCol.UpdateOne(ctx,
		filter, set,
		opts)
	return id, err
}

func (m *MongoDBArticleDao) SyncStatus(ctx context.Context, authorId, Id int64, status domain.ArticleStatus) error {
	filter := bson.D{bson.E{Key: "id", Value: Id},
		bson.E{Key: "author_id", Value: authorId}}
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.D{bson.E{Key: "status", Value: status}}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("ID 不对或者创作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, sets)
	return err
}

func NewArticleMongoDBDAO(node *snowflake.Node, mdb *mongo.Database) ArticleDAO {
	return &MongoDBArticleDao{
		node:    node,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
	}
}
