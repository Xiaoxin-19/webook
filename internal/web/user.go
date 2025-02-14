package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"time"
	"unicode/utf8"
	"webok/internal/domain"
	"webok/internal/service"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/ginx"
	"webok/pkg/logger"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	phoneRegexPattern    = `^1[3-9]\d{9}$`
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	phoneRexExp    *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
	log            logger.Logger
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwt ijwt.Handler, l logger.Logger) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		phoneRexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
		Handler:        jwt,
		log:            l,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WarpBody[SignUpReq](h.signUp))
	ug.POST("/login", ginx.WarpBody[LoginJwtReq](h.LoginJWT))
	ug.POST("/edit", ginx.WarpBodyAndClaims[EditReq, ijwt.TokenClaims](h.edit))
	ug.GET("/profile", ginx.WarpClaims[ijwt.TokenClaims](h.profile))
	ug.GET("/refresh_token", ginx.WarpClaims[ijwt.TokenClaims](h.ReFreshToken))
	ug.GET("/logout", ginx.Warp(h.logout))
	//验证码相关接口
	ug.POST("/login_sms/code/send", ginx.WarpBody[SendSMSReq](h.SendSMSLoginCode))
	ug.POST("/login_sms", ginx.WarpBody[LoginSMSReq](h.LoginSMS))
}

func (h *UserHandler) signUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		_ = ctx.Error(err)
	}
	if !isEmail {
		return ginx.Result{Code: 4, Msg: "邮箱格式错误"}, nil
	}

	if req.Password != req.ConfirmPassword {
		return ginx.Result{Code: 4, Msg: "两次密码不一致"}, nil
	}

	isPwd, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{}, err
	}
	if !isPwd {
		return ginx.Result{Code: 4, Msg: "密码格式错误"}, nil
	}

	err = h.svc.SignUp(ctx, &domain.User{Email: req.Email, Password: req.Password})
	switch {
	case err == nil:
		return ginx.Result{Msg: "注册成功"}, nil
	case errors.Is(err, service.ErrDuplicate):
		return ginx.Result{Code: 4, Msg: "邮箱冲突"}, nil
	default:
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
}

func (h *UserHandler) LoginJWT(ctx *gin.Context, req LoginJwtReq) (ginx.Result, error) {
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		err = h.SetLoginToken(ctx, u.Id)
		if err != nil {
			return ginx.Result{}, err
		}
		return ginx.Result{Msg: "登录成功"}, nil
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		return ginx.Result{Code: 4, Msg: "用户名或者密码不对"}, nil
	default:
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
}

func (h *UserHandler) ReFreshToken(ctx *gin.Context, uc ijwt.TokenClaims) (ginx.Result, error) {
	err := h.SetAccessToken(ctx, uc.Uid, uc.Ssid)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Msg: "刷新成功"}, nil
}

func (h *UserHandler) edit(ctx *gin.Context, req EditReq, uc ijwt.TokenClaims) (ginx.Result, error) {
	nicknameLen := utf8.RuneCountInString(req.NickName)
	if nicknameLen > 30 || nicknameLen <= 0 {
		return ginx.Result{Msg: "昵称的长度在0-30之间", Code: 4}, nil
	}
	t, err := time.Parse(time.DateOnly, req.BirthDay)
	if err != nil {
		return ginx.Result{Msg: "日期格式错误", Code: 4}, nil
	}

	briefLen := utf8.RuneCountInString(req.Brief)
	if briefLen > 256 {
		return ginx.Result{Code: 4, Msg: "简介的长度在0-256之间"}, nil
	}

	id := uc.Uid
	err = h.svc.ModifyNoSensitiveInfo(ctx, &domain.User{Id: id,
		Nickname: req.NickName,
		AboutMe:  req.Brief,
		Birthday: t})
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "修改成功"}, nil
}

func (h *UserHandler) profile(ctx *gin.Context, uc ijwt.TokenClaims) (ginx.Result, error) {
	id := uc.Uid
	u, err := h.svc.Profile(ctx, &domain.User{Id: id})

	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}

	type UserInfo struct {
		NickName string `json:"Nickname"`
		BirthDay string `json:"Birthday"`
		Phone    string `json:"Phone"`
		Email    string `json:"Email"`
		AboutMe  string `json:"AboutMe"`
	}

	return ginx.Result{Data: UserInfo{
		NickName: u.Nickname,
		BirthDay: u.Birthday.Format(time.DateOnly),
		AboutMe:  u.AboutMe,
		Email:    u.Email,
		Phone:    u.Phone,
	}}, nil
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context, req SendSMSReq) (ginx.Result, error) {

	ok, err := h.phoneRexExp.MatchString(req.Phone)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	if !ok {
		return ginx.Result{Code: 4, Msg: "手机格式不正确"}, nil
	}
	err = h.codeSvc.Send(ctx, bizLogin, req.Phone)

	switch {
	case err == nil:
		return ginx.Result{Msg: "发送成功"}, nil
	case errors.Is(err, service.ErrCodeSendTooMany):
		return ginx.Result{Code: 4, Msg: "短信发送太频繁，请稍后再试"}, nil
	default:
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {

	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	if !ok {
		return ginx.Result{Code: 4, Msg: "验证失败"}, nil
	}

	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	// 设置登录态
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "登录成功"}, nil
}

func (h *UserHandler) logout(ctx *gin.Context) (ginx.Result, error) {
	err := h.ClearToken(ctx)
	if err != nil {
		return ginx.Result{Msg: "系统错误", Code: 5}, err
	}
	return ginx.Result{Msg: "登出成功"}, nil
}
