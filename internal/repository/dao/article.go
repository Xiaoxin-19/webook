package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"webok/internal/domain"
)

var ErrUpdateFailed = errors.New("update failed, id - author_id not match")

type Article struct {
	ID       int64  `gorm:"primaryKey,autoIncrement" bson:"id, omitempty"`
	Title    string `gorm:"type:varchar(1024)" bson:"title, omitempty"`
	Content  string `gorm:"type:text" bson:"content, omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id, omitempty"`
	Status   uint8  `bson:"status, omitempty"`
	Ctime    int64  `bson:"ctime, omitempty"`
	Utime    int64  `bson:"utime, omitempty"`
}

type PublishedArticle Article

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, entity Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	SyncStatus(ctx context.Context, authorId, Id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

func (a *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	if err != nil {
		return PublishedArticle{}, err
	}
	return res, nil
}

func (a *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	if err != nil {
		return Article{}, err
	}
	return art, nil
}

func (a *ArticleGORMDAO) Sync(ctx context.Context, article Article) (int64, error) {
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
		pubArt := PublishedArticle(article)
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			// 对Mysql无效，但是可以兼容其他方言
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"status":  pubArt.Status,
				"utime":   now,
			}),
		}).Create(&pubArt).Error
		return err
	})

	if err != nil {
		return 0, err
	}
	return article.ID, nil
}

// SyncV1 自己管理事务
func (a *ArticleGORMDAO) SyncV1(ctx context.Context, article Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	dao := NewArticleGORMDAO(tx)

	var (
		id  int64
		err error
	)
	id = article.ID
	if id == 0 {
		id, err = dao.Insert(ctx, article)
	} else {
		err = dao.UpdateById(ctx, article)
	}
	if err != nil {
		return 0, err
	}

	article.ID = id
	now := time.Now().UnixMilli()
	pubArt := PublishedArticle(article)
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		// 对Mysql无效，但是可以兼容其他方言
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   now,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return article.ID, nil
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{db: db}
}

func (a *ArticleGORMDAO) Insert(ctx context.Context, article Article) (int64, error) {
	article.Ctime = time.Now().UnixMilli()
	article.Utime = article.Ctime
	err := a.db.WithContext(ctx).Create(&article).Error
	return article.ID, err
}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, entity Article) error {
	entity.Utime = time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&Article{}).Where("id = ? AND author_id = ?", entity.ID, entity.AuthorId).
		Updates(map[string]any{
			"title":   entity.Title,
			"content": entity.Content,
			"status":  entity.Status,
			"utime":   entity.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrUpdateFailed
	}
	return nil
}

func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, authorId, Id int64, status domain.ArticleStatus) error {
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

		return tx.Model(&PublishedArticle{}).Where("id = ?", Id).
			Updates(map[string]any{
				"status": status.ToUint8(),
				"utime":  now,
			}).Error
	})
	return err
}

func (a *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	arts := make([]Article, 0)
	err := a.db.WithContext(ctx).Where("author_id = ?", uid).
		Offset(offset).
		Limit(limit).
		Order("ctime DESC").
		Find(&arts).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return arts, nil
}
