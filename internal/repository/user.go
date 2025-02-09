package repository

import (
	"context"
	"database/sql"
	"errors"
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

//go:generate mockgen -source=user.go -package=repomocks -destination=./mock/user.mock.go
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	UpdateById(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindById(ctx context.Context, id int64) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func (ur *CachedUserRepository) Create(ctx context.Context, u *domain.User) error {
	return ur.dao.Insert(ctx, ur.toEntity(u))
}

func (ur *CachedUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return ur.toDomain(u), nil
}

func (ur *CachedUserRepository) FindById(ctx context.Context, id int64) (*domain.User, error) {
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

func (ur *CachedUserRepository) UpdateById(ctx context.Context, u *domain.User) error {
	err := ur.dao.UpdateById(ctx, ur.toEntity(u))
	if err != nil {
		return err
	}
	return nil
}

func NewCachedUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{dao: dao, cache: cache}
}

func (ur *CachedUserRepository) toDomain(u *dao.User) *domain.User {
	return &domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
		AboutMe:  u.AboutMe,
		Ctime:    time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionId.String,
		},
	}
}

func (ur *CachedUserRepository) toEntity(u *domain.User) *dao.User {
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
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
	}
}

func (ur *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u, err := ur.dao.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	return ur.toDomain(u), nil
}

func (repo *CachedUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	ue, err := repo.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return *repo.toDomain(&ue), nil
}
