package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"webok/internal/domain"
	"webok/internal/repository"
	"webok/pkg/logger"
)

//go:generate mockgen -source=Interactive.go -package=svcmocks -destination=./mock/Interactive.mock.go
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func (i *interactiveService) Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}

	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		intr.Liked, er = i.repo.Liked(ctx, biz, id, uid)
		return er
	})

	eg.Go(func() error {
		var er error
		intr.Collected, er = i.repo.Collected(ctx, biz, id, uid)
		return er
	})

	err = eg.Wait()
	if err != nil {
		i.l.Error("Get interactive failed",
			logger.Int64("id", id),
			logger.Int64("uid", uid),
			logger.Error(err))
	}
	return intr, nil
}

func (i *interactiveService) Collect(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := i.repo.AddCollectionItem(ctx, biz, id, cid, uid)
	return err
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.IncrLickCnt(ctx, biz, id, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLickCnt(ctx, biz, id, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository, log logger.Logger) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    log,
	}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	return i.repo.IncrReadCnt(ctx, biz, id)
}
