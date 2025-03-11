package repository

import (
	"context"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
)

//go:generate mockgen -source=Interactive.go -package=repomocks -destination=./mock/Interactive.mock.go
type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	IncrLickCnt(ctx context.Context, biz string, id int64, uid int64) error
	DecrLickCnt(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDao
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, biz, id, cid, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrCollectionCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrLickCnt(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.IncrLickCnt(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) DecrLickCnt(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DecrLickCnt(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
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
