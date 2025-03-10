package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webok/internal/domain"
	"webok/internal/repository"
)

var (
	ErrDuplicate             = repository.ErrDuplicate
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
	ErrRecordNotFound        = repository.ErrRecordNotFound
)

//go:generate mockgen -source=user.go -package=svcmocks -destination=./mock/user.mock.go
type UserService interface {
	SignUp(ctx context.Context, u *domain.User) error
	Login(ctx context.Context, email string, password string) (*domain.User, error)
	ModifyNoSensitiveInfo(ctx context.Context, u *domain.User) error
	Profile(ctx context.Context, d *domain.User) (*domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (*domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type NormalUserService struct {
	repo repository.UserRepository
}

func (us *NormalUserService) SignUp(ctx context.Context, u *domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return us.repo.Create(ctx, u)
}

func (us *NormalUserService) Login(ctx context.Context, email string, password string) (*domain.User, error) {
	u, err := us.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, ErrInvalidUserOrPassword
	}
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (us *NormalUserService) ModifyNoSensitiveInfo(ctx context.Context, u *domain.User) error {

	err := us.repo.UpdateById(ctx, u)
	if err != nil {
		return err
	}
	return nil
}

func (us *NormalUserService) Profile(ctx context.Context, d *domain.User) (*domain.User, error) {
	u, err := us.repo.FindById(ctx, d.Id)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (us *NormalUserService) FindOrCreate(ctx context.Context, phone string) (*domain.User, error) {
	u, err := us.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrRecordNotFound) {
		return u, err
	}

	err = us.repo.Create(ctx, &domain.User{Phone: phone})
	if err != nil && !errors.Is(err, ErrDuplicate) {
		return nil, err
	}
	// 如果有主从，则强制走主库，避免同步造成的读取失败
	return us.repo.FindByPhone(ctx, phone)
}

func (us *NormalUserService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := us.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if !errors.Is(err, repository.ErrRecordNotFound) {
		return u, err
	}

	err = us.repo.Create(ctx, &domain.User{WechatInfo: wechatInfo})
	if err != nil && !errors.Is(err, ErrDuplicate) {
		return domain.User{}, err
	}
	return us.repo.FindByWechat(ctx, wechatInfo.OpenId)
}

func NewNormalUserService(repo repository.UserRepository) UserService {
	return &NormalUserService{repo: repo}
}
