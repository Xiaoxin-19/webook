package repository

import (
	"context"
	"webok/internal/domain"
)

//go:generate mockgen -source=article_reader.go -package=repomocks -destination=./mock/article_reader.mock.go
type ArticleReaderRepository interface {
	// Save Create when article.Id is 0, otherwise update
	Save(ctx context.Context, art domain.Article) error
}
