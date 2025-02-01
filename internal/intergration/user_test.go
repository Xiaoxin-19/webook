package intergration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"webok/internal/intergration/startup"
	"webok/internal/web"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestUserHandler_SendSMS(t *testing.T) {
	server := startup.InitWebServer()
	rdb := startup.InitRedis()
	phone := "15212345678"
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		phone      string
		wantCode   int
		wantResult web.Result
	}{
		{
			name: "发送成功",
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
				assert.True(t, len(r) >= 6)

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
			phone:      phone,
			wantCode:   http.StatusOK,
			wantResult: web.Result{Msg: "发送成功"},
		},
		{
			name: "手机号码格式错误",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			phone:      "1234",
			wantCode:   http.StatusOK,
			wantResult: web.Result{Code: 4, Msg: "手机格式不正确"},
		},
		{
			name: "验证码发送频繁",
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
			phone:    phone,
			wantCode: http.StatusOK,
			wantResult: web.Result{Code: 4,
				Msg: "短信发送太频繁，请稍后再试"},
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
			phone:    phone,
			wantCode: http.StatusOK,
			wantResult: web.Result{
				Code: 5,
				Msg:  "系统错误"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.before(t)
			defer testCase.after(t)

			body := fmt.Sprintf(`{"phone":"%s"}`, testCase.phone)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", strings.NewReader(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			code := recorder.Result().StatusCode
			assert.Equal(t, testCase.wantCode, code)
			if code != http.StatusOK {
				// 用于测试Bind分支，当bind失败的时候，会返回401，而其他会返回200
				return
			}
			var res web.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, testCase.wantResult, res)
		})
	}
}
