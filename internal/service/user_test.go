package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"webok/internal/domain"
	"webok/internal/repository"
	repomocks "webok/internal/repository/mock"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("pass@1234")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, password)
	assert.NoError(t, err)
}

func TestNormalUserService_Login(t *testing.T) {
	testCase := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		email    string
		password string
		wantUser *domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(&domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$S38kdrCUDCeBbI39cf/8VOxlnD3nFk5AbVZgCGPOLQ3SDEyeOypFC",
						Phone:    "15212345678",
					}, nil)

				return repo
			},
			email:    "123@qq.com",
			password: "pass@1234",
			wantUser: &domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$S38kdrCUDCeBbI39cf/8VOxlnD3nFk5AbVZgCGPOLQ3SDEyeOypFC",
				Phone:    "15212345678",
			},
			wantErr: nil,
		},
		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(nil, repository.ErrRecordNotFound)

				return repo
			},
			email:    "123@qq.com",
			password: "pass@1234",
			wantUser: nil,
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(nil, errors.New("error"))

				return repo
			},
			email:    "123@qq.com",
			password: "pass@1234",
			wantUser: nil,
			wantErr:  errors.New("error"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(&domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$S38kdrCUDCeBbI39cf/8VOxlnD3nFk5AbVZgCGPOLQ3SDEyeOypFC",
						Phone:    "15212345678",
					}, nil)

				return repo
			},
			email:    "123@qq.com",
			password: "pass@124",
			wantUser: nil,
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			userSvc := NewNormalUserService(repo)
			gotUser, err := userSvc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantUser, gotUser)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
