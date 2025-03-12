package web

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
	"webok/internal/domain"
	"webok/internal/service"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/ginx"
	"webok/pkg/logger"
)

type ArticleHandler struct {
	log      logger.Logger
	svc      service.ArticleService
	interSvc service.InteractiveService
	biz      string
}

func NewArticleHandler(s service.ArticleService, l logger.Logger, isvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:      s,
		log:      l,
		interSvc: isvc,
		biz:      "article",
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
	pub.POST("/like", ginx.WarpBodyAndClaims[LikeArticleReq, ijwt.TokenClaims](h.like))
	pub.POST("/cancelLike", ginx.WarpBodyAndClaims[LikeArticleReq, ijwt.TokenClaims](h.like))
	pub.POST("/collect", ginx.WarpBodyAndClaims[CollectArticleReq, ijwt.TokenClaims](h.collect))

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
			Ctime:      time.UnixMilli(v.Ctime).Format(time.DateTime),
			Utime:      time.UnixMilli(v.Utime).Format(time.DateTime),
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
		Ctime:      time.UnixMilli(art.Ctime).Format(time.DateTime),
		Utime:      time.UnixMilli(art.Utime).Format(time.DateTime),
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

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)

	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx, id)
		if er != nil {
			h.log.Error("查询文章失败内容失败，", logger.String("id", idStr), logger.Error(err))
			return err
		}
		return nil
	})

	eg.Go(func() error {
		var er error
		intr, er = h.interSvc.Get(ctx, h.biz, id, uc.Uid)
		if er != nil {
			h.log.Error("查询文章交互状态失败", logger.String("id", idStr), logger.Error(err))
			return err
		}
		return nil
	})

	err = eg.Wait()
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := h.interSvc.IncrReadCnt(ctx, h.biz, art.Id)
		if er != nil {
			h.log.Error("增加阅读数失败", logger.Error(er), logger.Int64("id", art.Id))
		}
	}()

	artVo := ArticleVO{
		ID:         art.Id,
		Title:      art.Title,
		Content:    art.Content,
		AuthorId:   art.Author.Id,
		AuthorName: art.Author.Name,
		Status:     art.Status.ToUint8(),
		Ctime:      time.UnixMilli(art.Ctime).Format(time.DateTime),
		Utime:      time.UnixMilli(art.Utime).Format(time.DateTime),
		LikeCnt:    intr.LikeCnt,
		Collected:  intr.Collected,
		Liked:      intr.Liked,
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
	}
	return ginx.Result{
		Code: 0,
		Msg:  "",
		Data: artVo,
	}, err
}

func (h *ArticleHandler) like(ctx *gin.Context, req LikeArticleReq, claims ijwt.TokenClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		err = h.interSvc.Like(ctx, h.biz, req.Id, claims.Uid)
	} else {
		err = h.interSvc.CancelLike(ctx, h.biz, req.Id, claims.Uid)
	}
	if err != nil {
		h.log.Error("点赞失败", logger.Error(err), logger.Int64("uid", claims.Uid), logger.Int64("id", req.Id))
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "ok"}, nil
}

func (h *ArticleHandler) collect(ctx *gin.Context, req CollectArticleReq, claims ijwt.TokenClaims) (ginx.Result, error) {
	err := h.interSvc.Collect(ctx, h.biz, req.Id, req.Cid, claims.Uid)
	if err != nil {
		h.log.Error("收藏失败", logger.Error(err), logger.Int64("uid", claims.Uid), logger.Int64("id", req.Id))
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "ok"}, nil
}
