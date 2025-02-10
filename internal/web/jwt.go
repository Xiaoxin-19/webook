package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type JwtHandler struct {
}

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")
var JWTReFreshKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSprCwgK")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	ssid string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	ssid string
}

// SetAccessToken 设置 AccessToken
func (h *JwtHandler) SetAccessToken(ctx *gin.Context, userId int64) {
	uc := UserClaims{
		Uid: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
}

// SetLoginToken 设置登录 token,同时 AccessToken 和 RefreshToken
func (h *JwtHandler) SetLoginToken(ctx *gin.Context, userId int64) {
	h.SetAccessToken(ctx, userId)
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshClaims{
		Uid: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	})
	refreshToken, err := refresh.SignedString(JWTReFreshKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-refresh-token", refreshToken)
}

// ExtractToken 从请求头中提取 token
func (h *JwtHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	strArr := strings.Split(authCode, " ")
	if len(strArr) != 2 && strArr[0] != "Bearer" {
		return ""
	}
	return strArr[1]
}
