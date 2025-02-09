package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
	"webok/internal/web"
)

type LoginMiddlewareBuilder struct {
}

type LoginJWTMiddlewareBuilder struct {
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}

		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		strArr := strings.Split(authCode, " ")
		if len(strArr) != 2 && strArr[0] != "Bearer" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := strArr[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//if uc.UserAgent != ctx.GetHeader("User-Agent") {
		// 后期我们讲到了监控告警的时候，这个地方要埋点
		// 能够进来这个分支的，大概率是攻击者
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		expireTime := uc.ExpiresAt
		// 剩余过期时间 < 50s 就要刷新
		if expireTime.Sub(time.Now()) < time.Minute*10 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
			tokenStr, err = token.SignedString(web.JWTKey)
			ctx.Header("x-jwt-token", tokenStr)
			if err != nil {
				log.Println(err)
			}
		}
		ctx.Set("user", uc)
	}
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		var userId interface{}
		if userId = sess.Get("userId"); userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		const updateTimeKey = "updateTime"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Second*10 {
			// 你这是第一次进来
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			sess.Options(sessions.Options{
				// 十五分钟
				MaxAge: 30,
			})
			err := sess.Save()
			if err != nil {
				// 打日志
				fmt.Println(err)
			}
		}
	}
}
