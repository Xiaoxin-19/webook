package repository

import (
	"context"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
)

//go:generate mockgen -source=Interactive.go -package=repomocks -destination=./mock/Interactive.mock.go
type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDao
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, id)
	if err != nil {
		return err
	}
	return c.cache.IncrReadCntIfPresent(ctx, biz, id)
}

func NewCachedInteractiveRepository(dao dao.InteractiveDao, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}
