package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webok/internal/domain"
	"webok/internal/repository/cache"
	cachemocks "webok/internal/repository/cache/mock"
	"webok/internal/repository/dao"
	daomocks "webok/internal/repository/dao/mock"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO)
		ctx  context.Context
		uid  int64

		wantUser *domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存为命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(1)
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(nil, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(&dao.User{
					Id: 1,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "1231231",
					Birthday: 1000,
					AboutMe:  "about me",
					Phone: sql.NullString{
						String: "15212345678",
						Valid:  true,
					},
					Ctime: 123,
					Utime: 123,
				}, nil)
				c.EXPECT().Set(gomock.Any(), &domain.User{
					Id:       1,
					Email:    "123@qq.com",
					Password: "1231231",
					Birthday: time.UnixMilli(1000),
					AboutMe:  "about me",
					Phone:    "15212345678",
					Ctime:    time.UnixMilli(123),
				}).Return(nil)
				return c, d
			},
			ctx: context.Background(),
			uid: 1,
			wantUser: &domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "1231231",
				Birthday: time.UnixMilli(1000),
				AboutMe:  "about me",
				Phone:    "15212345678",
				Ctime:    time.UnixMilli(123),
			},
			wantErr: nil,
		},
		{
			name: "查找成功，缓存命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(1)
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(&domain.User{
					Id:       1,
					Email:    "123@qq.com",
					Password: "1231231",
					Birthday: time.UnixMilli(1000),
					AboutMe:  "about me",
					Phone:    "15212345678",
					Ctime:    time.UnixMilli(123),
				}, nil)
				return c, d
			},
			ctx: context.Background(),
			uid: 1,
			wantUser: &domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "1231231",
				Birthday: time.UnixMilli(1000),
				AboutMe:  "about me",
				Phone:    "15212345678",
				Ctime:    time.UnixMilli(123),
			},
			wantErr: nil,
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(1)
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(nil, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(nil, ErrRecordNotFound)
				return c, d
			},
			ctx:      context.Background(),
			uid:      1,
			wantUser: nil,
			wantErr:  ErrRecordNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(1)
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(nil, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).Return(&dao.User{
					Id: 1,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "1231231",
					Birthday: 1000,
					AboutMe:  "about me",
					Phone: sql.NullString{
						String: "15212345678",
						Valid:  true,
					},
					Ctime: 123,
					Utime: 123,
				}, nil)
				c.EXPECT().Set(gomock.Any(), &domain.User{
					Id:       1,
					Email:    "123@qq.com",
					Password: "1231231",
					Birthday: time.UnixMilli(1000),
					AboutMe:  "about me",
					Phone:    "15212345678",
					Ctime:    time.UnixMilli(123),
				}).Return(errors.New("redis error"))
				return c, d
			},
			ctx: context.Background(),
			uid: 1,
			wantUser: &domain.User{
				Id:       1,
				Email:    "123@qq.com",
				Password: "1231231",
				Birthday: time.UnixMilli(1000),
				AboutMe:  "about me",
				Phone:    "15212345678",
				Ctime:    time.UnixMilli(123),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userCache, userDao := tc.mock(ctrl)
			userRepo := NewCachedUserRepository(userDao, userCache)
			gotUser, err := userRepo.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, gotUser)
		})
	}

}
