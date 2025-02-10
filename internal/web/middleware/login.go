package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
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
