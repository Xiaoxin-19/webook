package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	longUUID "github.com/google/uuid"
	uuid "github.com/lithammer/shortuuid/v4"
	"webok/internal/service"
	"webok/internal/service/outh2/wechat"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/ginx"
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
	g.GET("/authurl", ginx.Warp(o.Auth2URL))
	g.Any("/callback", ginx.Warp(o.Callback)) // 不知道微信到底会返回一个什么请求
}

func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) (ginx.Result, error) {
	val, err := o.svc.AuthURL(ctx)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	state := uuid.New()
	err = o.setStateCookie(ctx, state)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Data: val}, nil
}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) (ginx.Result, error) {
	err := o.verifyState(ctx)
	if err != nil {
		return ginx.Result{Msg: "非法请求", Code: 4}, err
	}
	// 你校验不校验都可以
	code := ctx.Query("code")
	// state := ctx.Query("state")
	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		return ginx.Result{Msg: "授权码有误", Code: 4}, err
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	// 设置登录态
	ssid := longUUID.New().String()
	if err := o.SetAccessToken(ctx, u.Id, ssid); err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "Ok"}, nil

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
