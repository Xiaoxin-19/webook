package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webok/internal/domain"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, authorId int64, authorList []domain.Article) error
	DelFirstPage(ctx context.Context, authorId int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, res domain.Article) error
	DelPub(ctx context.Context, id int64) error
}

type ArticleRedisCache struct {
	cmd redis.Cmdable
}

func (a *ArticleRedisCache) DelPub(ctx context.Context, id int64) error {
	return a.cmd.Del(ctx, a.pubArticleKey(id)).Err()
}

func NewArticleRedisCache(cmd redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{cmd: cmd}
}

func (a *ArticleRedisCache) firstPageKey(authorId int64) string {
	return fmt.Sprintf("article:first_page:%d", authorId)
}

func (a *ArticleRedisCache) articleKey(id int64) string {
	return fmt.Sprintf("article:%d", id)
}
func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, authorId int64) ([]domain.Article, error) {
	val, err := a.cmd.Get(ctx, a.firstPageKey(authorId)).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, authorId int64, authorList []domain.Article) error {
	for i, _ := range authorList {
		authorList[i].Content = authorList[i].Abstract()
	}
	key := a.firstPageKey(authorId)
	data, err := json.Marshal(authorList)
	if err != nil {
		return err
	}
	return a.cmd.Set(ctx, key, data, 10*time.Minute).Err()
}

func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, authorId int64) error {
	return a.cmd.Del(ctx, a.firstPageKey(authorId)).Err()
}

func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	key := a.articleKey(id)
	val, err := a.cmd.Get(ctx, key).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) Set(ctx context.Context, art domain.Article) error {
	key := a.articleKey(art.Id)
	data, err := json.Marshal(art)
	if err != nil {
		// log
		return err
	}
	return a.cmd.Set(ctx, key, data, 10*time.Minute).Err()
}

func (a *ArticleRedisCache) pubArticleKey(id int64) string {
	return fmt.Sprintf("article:pub:%d", id)
}

func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {

	res, err := a.cmd.Get(ctx, a.pubArticleKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(res, &art)
	if err != nil {
		return domain.Article{}, err
	}
	return art, nil
}

func (a *ArticleRedisCache) SetPub(ctx context.Context, res domain.Article) error {
	key := a.pubArticleKey(res.Id)
	err := a.cmd.Set(ctx, key, res, 10*time.Minute).Err()
	return err
}
