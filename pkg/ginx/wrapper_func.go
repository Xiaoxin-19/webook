package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"webok/pkg/logger"
)

var L logger.Logger = logger.NewNopLogger()

func WarpBodyAndClaims[Req any, Claims jwt.Claims](Func func(ctx *gin.Context, req Req, claims Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			L.Error("bind request failed", logger.Field{
				Key: "error",
				Val: err,
			})
			return
		}

		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		resp, err := Func(ctx, req, uc)
		if err != nil {
			L.Error("handle request failed", logger.Field{
				Key: "error",
				Val: err,
			})
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

func WarpBody[Req any](Func func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			L.Error("bind request failed", logger.Field{
				Key: "error",
				Val: err,
			})
			return
		}

		resp, err := Func(ctx, req)
		if err != nil {
			L.Error("handle request failed", logger.Field{
				Key: "error",
				Val: err,
			})
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

func WarpClaims[Claims jwt.Claims](Func func(ctx *gin.Context, claims Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		resp, err := Func(ctx, uc)
		if err != nil {
			L.Error("handle request failed", logger.Field{
				Key: "error",
				Val: err,
			})
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

func Warp(Func func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := Func(ctx)
		if err != nil {
			L.Error("handle request failed", logger.Field{
				Key: "error",
				Val: err,
			})
		}
		ctx.JSON(http.StatusOK, resp)
	}
}
