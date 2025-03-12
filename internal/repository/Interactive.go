package repository

import (
	"context"
	"errors"
	"webok/internal/domain"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
	"webok/pkg/logger"
)

//go:generate mockgen -source=Interactive.go -package=repomocks -destination=./mock/Interactive.mock.go
type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	IncrLickCnt(ctx context.Context, biz string, id int64, uid int64) error
	DecrLickCnt(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDao
	cache cache.InteractiveCache
	l     logger.Logger
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikedInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectionInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, id)

	if err == nil {
		return intr, nil
	}

	ie, err := c.dao.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}
	res := c.toDomain(ie)
	err = c.cache.Set(ctx, biz, id, res)
	if err != nil {
		c.l.Error("cache set failed",
			logger.Int64("id", id),
			logger.String("biz", biz),
			logger.Error(err))
	}
	return res, nil
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

func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
		Liked:      false,
		Collected:  false,
	}
}

func NewCachedInteractiveRepository(dao dao.InteractiveDao, cache cache.InteractiveCache, log logger.Logger) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     log,
	}
}
