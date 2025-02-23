package repository

import (
	"context"
	"webok/internal/domain"
)

//go:generate mockgen -source=article_author.go -package=repomocks -destination=./mock/article_author.mock.go
type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}
