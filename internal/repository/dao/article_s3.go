package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
	"webok/internal/domain"
)

const bucket = "webook-1303500761"

type PublishedArticleS3 struct {
	ID       int64  `gorm:"primaryKey,autoIncrement" bson:"id, omitempty"`
	Title    string `gorm:"type:varchar(1024)" bson:"title, omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id, omitempty"`
	Status   uint8  `bson:"status, omitempty"`
	Ctime    int64  `bson:"ctime, omitempty"`
	Utime    int64  `bson:"utime, omitempty"`
}
type ArticleS3DAO struct {
	ArticleGORMDAO
	s3 *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, s3 *s3.S3) ArticleDAO {
	return &ArticleS3DAO{ArticleGORMDAO: ArticleGORMDAO{db: db}, s3: s3}
}
func (a *ArticleS3DAO) Sync(ctx context.Context, article Article) (int64, error) {
	var (
		id  int64
		err error
	)
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := NewArticleGORMDAO(tx)

		id = article.ID
		if id == 0 {
			id, err = dao.Insert(ctx, article)
		} else {
			err = dao.UpdateById(ctx, article)
		}
		if err != nil {
			return err
		}

		article.ID = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticleS3{
			ID:       article.ID,
			Title:    article.Title,
			AuthorId: article.AuthorId,
			Status:   article.Status,
			Ctime:    now,
			Utime:    now,
		}
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			// 对Mysql无效，但是可以兼容其他方言
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":  pubArt.Title,
				"status": pubArt.Status,
				"utime":  now,
			}),
		}).Create(&pubArt).Error
		return err
	})

	if err != nil {
		return 0, err
	}

	_, err = a.s3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string](bucket),
		Key:         ekit.ToPtr[string](strconv.FormatInt(article.ID, 10)),
		Body:        bytes.NewReader([]byte(article.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})

	return article.ID, err
}

func (a *ArticleS3DAO) SyncStatus(ctx context.Context, authorId, Id int64, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Article{}).Where("id = ? AND author_id = ?", Id, authorId).
			Updates(map[string]any{
				"status": status.ToUint8(),
				"utime":  now,
			})
		if err.Error != nil {
			return err.Error
		}
		if err.RowsAffected != 1 {
			return errors.New("ID and author_id not match")
		}

		return tx.Model(&PublishedArticleS3{}).Where("id = ?", Id).
			Updates(map[string]any{
				"status": status.ToUint8(),
				"utime":  now,
			}).Error
	})

	const statusPrivate = 3
	if status == statusPrivate {
		_, err = a.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string](bucket),
			Key:    ekit.ToPtr[string](strconv.FormatInt(Id, 10)),
		})
	}
	return err
}
