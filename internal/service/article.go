package service

import (
	"context"
	"webok/internal/domain"
	"webok/internal/repository"
	"webok/pkg/logger"
)

//go:generate mockgen -source=article.go -package=svcmocks -destination=./mock/article.mock.go
type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, articleId int64) error
}

type articleService struct {
	repo repository.ArticleRepository
	l    logger.Logger

	// v1 specific
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
}

func (a *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, article)
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(reader repository.ArticleReaderRepository, author repository.ArticleAuthorRepository, l logger.Logger) *articleService {
	return &articleService{
		readerRepo: reader,
		authorRepo: author,
		l:          l,
	}
}
func (a *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	var (
		id  int64
		err error
	)
	id = article.Id
	if id > 0 {
		err = a.authorRepo.Update(ctx, article)
	} else {
		id, err = a.authorRepo.Create(ctx, article)
	}
	if err != nil {
		return id, err
	}
	article.Id = id
	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, article)

		if err == nil {

			return id, nil
		}
		a.l.Error("部分失败，保存到线上库,重试", logger.Field{
			Key: "retry_count",
			Val: i,
		}, logger.Field{
			Key: "id",
			Val: id,
		}, logger.Error(err))
	}
	a.l.Error("部分失败，保存到线上库失败,重试3次失败", logger.Field{
		Key: "id",
		Val: id,
	}, logger.Error(err))

	return 0, err
}

func (a *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	if article.Id > 0 {
		return article.Id, a.repo.Update(ctx, article)
	}
	return a.repo.Create(ctx, article)
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, articleId int64) error {
	return a.repo.SyncStatus(ctx, uid, articleId)
}
