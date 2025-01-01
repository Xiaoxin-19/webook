package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"webok/internal/domain"
	"webok/internal/repository"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
	ErrRecordNotFound        = repository.ErrRecordNotFound
)

type UserService struct {
	repo *repository.UserRepository
}

func (us *UserService) SignUp(ctx context.Context, u *domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return us.repo.Create(ctx, u)
}

func (us *UserService) Login(ctx *gin.Context, email string, password string) (*domain.User, error) {
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

func (us *UserService) ModifyNoSensitiveInfo(ctx context.Context, u *domain.User) error {

	err := us.repo.UpdateById(ctx, u)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserService) Profile(ctx *gin.Context, d *domain.User) (*domain.User, error) {
	u, err := us.repo.FindById(ctx, int64(d.Id))
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, ErrInvalidUserOrPassword
	}
	return u, nil
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}
