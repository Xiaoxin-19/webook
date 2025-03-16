package service

import (
	"context"
	"webok/internal/domain"
	"webok/internal/events/article"
	"webok/internal/repository"
	"webok/pkg/logger"
)

//go:generate mockgen -source=article.go -package=svcmocks -destination=./mock/article.mock.go
type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, articleId int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository
	l    logger.Logger

	// v1 specific
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
	producer   article.Producer
}

func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			// 在这里发一个消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(err))
			}
		}
	}()

	return res, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, article)
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func NewArticleServiceV1(reader repository.ArticleReaderRepository, author repository.ArticleAuthorRepository, l logger.Logger, producer article.Producer) *articleService {
	return &articleService{
		readerRepo: reader,
		authorRepo: author,
		l:          l,
		producer:   producer,
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

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}
