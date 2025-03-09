package web

import (
	"github.com/gin-gonic/gin"
	"strconv"
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
	ug.POST("/list", ginx.WarpBodyAndClaims[Page, ijwt.TokenClaims](h.list))
	ug.GET("/detail/:id", ginx.WarpClaims[ijwt.TokenClaims](h.detail))
	pub := ug.Group("/pub")
	pub.GET("/:id", ginx.WarpClaims[ijwt.TokenClaims](h.pubDetail))
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

func (h *ArticleHandler) list(ctx *gin.Context, req Page, uc ijwt.TokenClaims) (ginx.Result, error) {
	list, err := h.svc.GetByAuthor(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		h.log.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", req.Offset),
			logger.Int("limit", req.Limit),
			logger.Int64("uid", uc.Uid))
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	data := make([]ArticleVO, len(list))
	for i, v := range list {
		vo := ArticleVO{
			ID:    v.Id,
			Title: v.Title,
			//Content:    v.Content,
			Abstract:   v.Abstract(),
			AuthorId:   v.Author.Id,
			AuthorName: v.Author.Name,
			Status:     v.Status.ToUint8(),
			Ctime:      v.Ctime,
			Utime:      v.Utime,
		}
		data[i] = vo
	}
	return ginx.Result{Data: data}, nil
}

func (h *ArticleHandler) detail(ctx *gin.Context, uc ijwt.TokenClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warn("查询文章失败，id 格式不对",
			logger.String("id", idStr),
			logger.Error(err))
		return ginx.Result{Msg: "参数错误", Code: 4}, err
	}

	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		h.log.Error("查询文章失败",
			logger.Error(err),
			logger.Int64("id", id))
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	if art.Author.Id != uc.Uid {
		h.log.Warn("查询文章失败，没有权限", logger.Int64("id", id), logger.Int64("uid", uc.Uid))
		return ginx.Result{Msg: "系统错误", Code: 5}, nil
	}

	vo := ArticleVO{
		ID:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		//Abstract:   art.Abstract(),
		AuthorId:   art.Author.Id,
		AuthorName: art.Author.Name,
		Status:     art.Status.ToUint8(),
		Ctime:      art.Ctime,
		Utime:      art.Utime,
	}
	return ginx.Result{Data: vo}, nil
}

func (h *ArticleHandler) pubDetail(ctx *gin.Context, uc ijwt.TokenClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warn("查询文章失败，id 格式不对",
			logger.String("id", idStr),
			logger.Error(err))
		return ginx.Result{Msg: "参数错误", Code: 4}, err
	}

	art, err := h.svc.GetPubById(ctx, id)
	if err != nil {
		h.log.Error("查询文章失败， 系统错误", logger.String("id", idStr), logger.Error(err))
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}

	artVo := ArticleVO{
		ID:         art.Id,
		Title:      art.Title,
		Content:    art.Content,
		AuthorId:   art.Author.Id,
		AuthorName: art.Author.Name,
		Status:     art.Status.ToUint8(),
		Ctime:      art.Ctime,
		Utime:      art.Utime,
	}
	return ginx.Result{
		Code: 0,
		Msg:  "",
		Data: artVo,
	}, err
}
