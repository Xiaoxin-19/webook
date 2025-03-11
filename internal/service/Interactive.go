package service

import (
	"context"
	"webok/internal/repository"
)

//go:generate mockgen -source=Interactive.go -package=svcmocks -destination=./mock/Interactive.mock.go
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, id int64, cid int64, uid int64) error
}

type interactiveService struct {
	repo repository.InteractiveRepository
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

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	return i.repo.IncrReadCnt(ctx, biz, id)
}
