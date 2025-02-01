package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"webok/internal/repository"
	"webok/internal/service/sms"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

//go:generate mockgen -source=code.go -package=svcmocks -destination=./mock/code.mock.go
type CodeService interface {
	generate() string
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
	Send(ctx context.Context, biz, phone string) error
}

type NormalCodeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func (svc *NormalCodeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	// 你在这儿，是不是要开始发送验证码了？
	if err != nil {
		return err
	}
	// 很少改，可以使用模板
	const codeTplId = "1877556"
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *NormalCodeService) Verify(ctx context.Context,
	biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if errors.Is(err, repository.ErrCodeVerifyTooMany) {
		// 相当于，我们对外面屏蔽了验证次数过多的错误，我们就是告诉调用者，你这个不对
		return false, nil
	}
	return ok, err
}

func (svc *NormalCodeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &NormalCodeService{
		repo: repo,
		sms:  smsSvc,
	}
}
