package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"webok/internal/domain"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
)

var (
	ErrDuplicate      = dao.ErrDuplicateEmail
	ErrRecordNotFound = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func (ur *UserRepository) Create(ctx context.Context, u *domain.User) error {
	return ur.dao.Insert(ctx, ur.toEntity(u))
}

func (ur *UserRepository) FindByEmail(ctx *gin.Context, email string) (*domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return ur.toDomain(u), nil
}

func (ur *UserRepository) FindById(ctx *gin.Context, id int64) (*domain.User, error) {
	du, err := ur.cache.Get(ctx, id)
	// 只要 err 为 nil，就返回
	switch {
	case err == nil:
		return du, nil
	case errors.Is(err, cache.ErrKeyNotExist):
		u, err := ur.dao.FindById(ctx, id)
		if err != nil {
			return nil, err
		}
		du = ur.toDomain(u)
		err = ur.cache.Set(ctx, du)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
			log.Println(err)
		}
		return du, nil
	default:
		// 接近降级的写法
		log.Printf("缓存查询失败，直接返回，不进行数据库查询， err:%v", err)
		return nil, err
	}
}

func (ur *UserRepository) UpdateById(ctx context.Context, u *domain.User) error {
	err := ur.dao.UpdateById(ctx, ur.toEntity(u))
	if err != nil {
		return err
	}
	return nil
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{dao: dao, cache: cache}
}

func (ur *UserRepository) toDomain(u *dao.User) *domain.User {
	return &domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
		AboutMe:  u.AboutMe,
	}
}

func (ur *UserRepository) toEntity(u *domain.User) *dao.User {
	return &dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
	}
}

func (ur *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u, err := ur.dao.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	return ur.toDomain(u), nil
}
