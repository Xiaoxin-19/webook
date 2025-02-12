package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	longUUID "github.com/google/uuid"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"webok/internal/service"
	"webok/internal/service/outh2/wechat"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/logger"
)

type OAuth2WechatHandler struct {
	ijwt.Handler
	svc             wechat.Service
	userSvc         service.UserService
	key             []byte
	stateCookieName string
	log             logger.Logger
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, jwt ijwt.Handler, l logger.Logger) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgB"),
		stateCookieName: "jwt-state",
		Handler:         jwt,
		log:             l,
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.Auth2URL)
	// 不知道微信到底会返回一个什么请求
	g.Any("/callback", o.Callback)
}

func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	val, err := o.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造授权链接失败",
		})
		return
	}
	state := uuid.New()
	err = o.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})
}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	// 你校验不校验都可以
	code := ctx.Query("code")
	// state := ctx.Query("state")
	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	// 设置登录态
	ssid := longUUID.New().String()
	o.SetAccessToken(ctx, u.Id, ssid)
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
	return

}
func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(o.key)
	if err != nil {
		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil
}
