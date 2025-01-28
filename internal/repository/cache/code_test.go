package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	redismocks "webok/internal/repository/cache/redismock"
)

func TestCodeRedisCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}

	testCases := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) redis.Cmdable
		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "set 成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				resCmd := redis.NewCmd(context.Background())
				resCmd.SetVal(int64(0))
				resCmd.SetErr(nil)
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode,
						[]string{keyFunc("test", "12312345678")}, []any{"123456"}).
					Return(resCmd)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "12312345678",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis 返回error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				resCmd := redis.NewCmd(context.Background())
				resCmd.SetVal(int64(0))
				resCmd.SetErr(errors.New("redis error"))
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode,
						[]string{keyFunc("test", "12312345678")}, []any{"123456"}).
					Return(resCmd)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "12312345678",
			code:    "123456",
			wantErr: errors.New("redis error"),
		},
		{
			name: "验证码不存在过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				resCmd := redis.NewCmd(context.Background())
				resCmd.SetVal(int64(-2))
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode,
						[]string{keyFunc("test", "12312345678")}, []any{"123456"}).
					Return(resCmd)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "12312345678",
			code:    "123456",
			wantErr: errors.New("验证码存在，但是没有过期时间"),
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				resCmd := redis.NewCmd(context.Background())
				resCmd.SetVal(int64(-1))
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode,
						[]string{keyFunc("test", "12312345678")}, []any{"123456"}).
					Return(resCmd)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "12312345678",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cmd := tc.mock(ctrl)
			c := NewCodeRedisCache(cmd)
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)

		})
	}
}
