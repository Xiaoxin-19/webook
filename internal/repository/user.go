package repository

import (
	"context"
	"github.com/gin-gonic/gin"
	"webok/internal/domain"
	"webok/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrRecordNotFound = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func (ur *UserRepository) Create(ctx context.Context, u *domain.User) error {
	return ur.dao.Insert(ctx, &dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (ur *UserRepository) FindByEmail(ctx *gin.Context, email string) (*domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return ur.toDomain(u), nil
}

func (ur *UserRepository) toDomain(u *dao.User) *domain.User {
	return &domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Brief:    u.Brief,
	}
}

func (ur *UserRepository) FindById(ctx *gin.Context, id int64) (*domain.User, error) {
	u, err := ur.dao.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	return ur.toDomain(u), nil
}

func (ur *UserRepository) UpdateById(ctx context.Context, u *domain.User) error {
	err := ur.dao.UpdateById(ctx, &dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Brief:    u.Brief,
	})
	if err != nil {
		return err
	}
	return nil
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: dao}
}
