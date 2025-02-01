package cache_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"webok/internal/intergration/startup"
	"webok/internal/repository/cache"
)

func TestRedisLuaCodeCache_Set_e2e(t *testing.T) {

	rdb := startup.InitRedis()
	phone := "15212345678"
	testCase := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "验证码设置成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()

				// 验证值是否存在
				keyNumber := fmt.Sprintf("%s:%s:%s", "phone_code", "login", phone)
				t.Log(keyNumber)
				r, err := rdb.Get(ctx, keyNumber).Result()
				assert.NoError(t, err)
				assert.True(t, r == "123456")

				keyCnt := fmt.Sprintf("%s:%s", keyNumber, "cnt")
				r, err = rdb.Get(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, r == "3")

				//验证过期时间
				ttl, err := rdb.TTL(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, ttl > time.Minute*9+time.Second*50)

				ttl, err = rdb.TTL(ctx, keyNumber).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, ttl > time.Minute*9+time.Second*50)

				//清除值
				_, err = rdb.Del(ctx, keyNumber).Result()
				assert.NoError(t, err)
				_, err = rdb.Del(ctx, keyCnt).Result()
				assert.NoError(t, err)
			},
			phone: phone,
			code:  "123456",
			biz:   "login",
			ctx:   context.Background(),
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				// 设置redis数据
				keyNumber := fmt.Sprintf("%s:%s:%s", "phone_code", "login", phone)
				_, err := rdb.Set(ctx, keyNumber, "123456", time.Minute*9+time.Second*50).Result()
				assert.NoError(t, err)

				keyCnt := fmt.Sprintf("%s:%s", keyNumber, "cnt")
				_, err = rdb.Set(ctx, keyCnt, "3", time.Minute*9+time.Second*50).Result()

				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()

				// 验证值是否存在
				keyNumber := fmt.Sprintf("%s:%s:%s", "phone_code", "login", phone)
				t.Log(keyNumber)
				r, err := rdb.Get(ctx, keyNumber).Result()
				assert.NoError(t, err)
				assert.True(t, r == "123456")

				keyCnt := fmt.Sprintf("%s:%s", keyNumber, "cnt")
				r, err = rdb.Get(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, r == "3")

				//验证过期时间
				ttl, err := rdb.TTL(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, ttl > time.Minute*9+time.Second*30)

				ttl, err = rdb.TTL(ctx, keyNumber).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, ttl > time.Minute*9+time.Second*30)

				//清除值
				_, err = rdb.Del(ctx, keyNumber).Result()
				assert.NoError(t, err)
				_, err = rdb.Del(ctx, keyCnt).Result()
				assert.NoError(t, err)
			},
			phone:   phone,
			code:    "123456",
			biz:     "login",
			ctx:     context.Background(),
			wantErr: cache.ErrCodeSendTooMany,
		},
		{
			name: "未知错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				// 设置redis数据
				keyNumber := fmt.Sprintf("%s:%s:%s", "phone_code", "login", phone)
				_, err := rdb.Set(ctx, keyNumber, "123456", 0).Result()
				assert.NoError(t, err)

				keyCnt := fmt.Sprintf("%s:%s", keyNumber, "cnt")
				_, err = rdb.Set(ctx, keyCnt, "3", 0).Result()

				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()

				// 验证值是否存在
				keyNumber := fmt.Sprintf("%s:%s:%s", "phone_code", "login", phone)
				t.Log(keyNumber)
				r, err := rdb.Get(ctx, keyNumber).Result()
				assert.NoError(t, err)
				assert.True(t, r == "123456")

				keyCnt := fmt.Sprintf("%s:%s", keyNumber, "cnt")
				r, err = rdb.Get(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, true, r == "3")

				//清除值
				_, err = rdb.Del(ctx, keyNumber).Result()
				assert.NoError(t, err)
				_, err = rdb.Del(ctx, keyCnt).Result()
				assert.NoError(t, err)
			},
			phone:   phone,
			code:    "123456",
			biz:     "login",
			ctx:     context.Background(),
			wantErr: cache.ErrCodeNoExpireTime,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			c := cache.NewCodeRedisCache(rdb)
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
