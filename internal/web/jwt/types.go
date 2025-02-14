package jwt

import "github.com/gin-gonic/gin"

//go:generate mockgen -destination ./mock/jwt.mock.go -package jwtmocks -source types.go
type Handler interface {
	SetAccessToken(ctx *gin.Context, userId int64, ssid string) error
	SetLoginToken(ctx *gin.Context, userId int64) error
	SetRefreshToken(ctx *gin.Context, userId int64, ssid string) error
	ExtractToken(ctx *gin.Context) string
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) error
	ParseAccessToken(tokenStr string) (*TokenClaims, error)
	ParseRefreshToken(tokenStr string) (*TokenClaims, error)
}
