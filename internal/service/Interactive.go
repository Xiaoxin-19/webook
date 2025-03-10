package service

import (
	"context"
	"webok/internal/repository"
)

//go:generate mockgen -source=Interactive.go -package=svcmocks -destination=./mock/Interactive.mock.go
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	return i.repo.IncrReadCnt(ctx, biz, id)
}
