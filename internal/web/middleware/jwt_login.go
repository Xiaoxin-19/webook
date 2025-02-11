package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	ijwt "webok/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
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

		tokenStr := m.ExtractToken(ctx)
		var uc *ijwt.TokenClaims
		var err error
		if path == "/users/refresh_token" {
			uc, err = m.ParseRefreshToken(tokenStr)
		} else {
			uc, err = m.ParseAccessToken(tokenStr)
		}

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("user", *uc)
	}
}
