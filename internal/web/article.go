package web

import (
	"github.com/gin-gonic/gin"
	"webok/internal/domain"
	"webok/internal/service"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/ginx"
	"webok/pkg/logger"
)

type ArticleHandler struct {
	log logger.Logger
	svc service.ArticleService
}

func NewArticleHandler(s service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: s,
		log: l,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/articles")

	ug.POST("/edit", ginx.WarpBodyAndClaims[EditArticleReq, ijwt.TokenClaims](h.edit))
	ug.POST("/publish", ginx.WarpBodyAndClaims[PublishArticleReq, ijwt.TokenClaims](h.publish))
	ug.POST("/withdraw", ginx.WarpBodyAndClaims[WithdrawArticleReq, ijwt.TokenClaims](h.withdraw))
}

// edit 编辑文章 返回文章ID
func (h *ArticleHandler) edit(ctx *gin.Context, req EditArticleReq, uc ijwt.TokenClaims) (ginx.Result, error) {
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}

	return ginx.Result{Data: id}, nil
}

func (h *ArticleHandler) publish(ctx *gin.Context, req PublishArticleReq, uc ijwt.TokenClaims) (ginx.Result, error) {
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}

	return ginx.Result{Data: id, Msg: "发布成功"}, nil
}

func (h *ArticleHandler) withdraw(ctx *gin.Context, req WithdrawArticleReq, claims ijwt.TokenClaims) (ginx.Result, error) {
	err := h.svc.Withdraw(ctx, claims.Uid, req.Id)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "Ok"}, nil
}
