package jwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisHandler struct {
	rdb                 redis.Cmdable
	accessTokenKey      []byte
	refreshTokenKey     []byte
	refreshExpirationAt time.Duration
	keyPrefix           string
}

func NewRedisHandler(rdb redis.Cmdable) Handler {
	return &RedisHandler{
		rdb:                 rdb,
		accessTokenKey:      []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK"),
		refreshTokenKey:     []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgA"),
		refreshExpirationAt: time.Hour * 24 * 7,
		keyPrefix:           "users:ssid:",
	}
}

type TokenClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

// SetAccessToken 设置 AccessToken
func (h *RedisHandler) SetAccessToken(ctx *gin.Context, userId int64, ssid string) error {
	uc := TokenClaims{
		Uid:  userId,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(h.accessTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// SetLoginToken 设置登录 token,同时 AccessToken 和 RefreshToken
func (h *RedisHandler) SetLoginToken(ctx *gin.Context, userId int64) error {
	ssid := uuid.New().String()
	err := h.SetRefreshToken(ctx, userId, ssid)
	if err != nil {
		return nil
	}
	return h.SetAccessToken(ctx, userId, ssid)
}

func (h *RedisHandler) SetRefreshToken(ctx *gin.Context, userId int64, ssid string) error {
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, TokenClaims{
		Uid:  userId,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.refreshExpirationAt)),
		},
	})
	refreshToken, err := refresh.SignedString(h.refreshTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", refreshToken)
	return nil
}

// ExtractToken 从请求头中提取 token
func (h *RedisHandler) ExtractToken(ctx *gin.Context) string {
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

func (h *RedisHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(TokenClaims)
	strBuilder := strings.Builder{}
	strBuilder.WriteString(h.keyPrefix)
	strBuilder.WriteString(uc.Ssid)
	return h.rdb.Set(ctx, strBuilder.String(), "", time.Hour*24).Err()
}

func (h *RedisHandler) CheckSession(ctx *gin.Context, ssid string) error {
	strBuilder := strings.Builder{}
	strBuilder.WriteString(h.keyPrefix)
	strBuilder.WriteString(ssid)
	cnt, err := h.rdb.Exists(ctx, strBuilder.String()).Result()
	// 这里可以做降级处理，防止redis奔溃导致服务奔溃
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (h *RedisHandler) ParseAccessToken(tokenStr string) (*TokenClaims, error) {
	return h.parseToken(tokenStr, h.accessTokenKey)
}

func (h *RedisHandler) ParseRefreshToken(tokenStr string) (*TokenClaims, error) {
	return h.parseToken(tokenStr, h.refreshTokenKey)
}

func (h *RedisHandler) parseToken(tokenStr string, key []byte) (*TokenClaims, error) {
	var uc TokenClaims
	token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, errors.New("token 无效")
	}
	return &uc, nil
}
