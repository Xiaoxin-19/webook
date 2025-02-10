package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"webok/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
	web.JwtHandler
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
		var uc web.UserClaims
		var token *jwt.Token
		var err error
		if path == "/users/refresh_token" {
			token, err = jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
				return web.JWTReFreshKey, nil
			})
		} else {
			token, err = jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
				return web.JWTKey, nil
			})
		}

		if err != nil || token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("user", uc)
	}
}
