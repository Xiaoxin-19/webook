package dao

import (
	"context"
	"gorm.io/gorm"
)

//go:generate mockgen -source=article_author.go -package=daomocks -destination=./mock/article_author.mock.go
type ArticleAuthorDAO interface {
	Create(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
}

type ArticleAuthorGORMDAO struct {
	db *gorm.DB
}

func NewArticleAuthorGORMDAO(db *gorm.DB) ArticleAuthorDAO {
	return &ArticleAuthorGORMDAO{db: db}
}

func (a *ArticleAuthorGORMDAO) Create(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (a *ArticleAuthorGORMDAO) UpdateById(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}
