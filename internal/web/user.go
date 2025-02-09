package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
	"unicode/utf8"
	"webok/internal/domain"
	"webok/internal/service"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	phoneRegexPattern    = `^1[3-9]\d{9}$`
	bizLogin             = "login"
)

type UserHandler struct {
	jwtHandler
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	phoneRexExp    *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		phoneRexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", h.signUp)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.edit)
	ug.GET("/profile", h.profile)
	//验证码相关接口
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)
}

func (h *UserHandler) signUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式不符合")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不匹配")
		return
	}

	isPwd, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPwd {
		ctx.String(http.StatusOK, "密码格式错误")
		return
	}

	err = h.svc.SignUp(ctx, &domain.User{Email: req.Email, Password: req.Password})
	switch {
	case err == nil:
		ctx.String(http.StatusOK, "注册成功")
	case errors.Is(err, service.ErrDuplicate):
		ctx.String(http.StatusOK, "已注册,请登录")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		h.setJwtToken(ctx, u.Id)
		ctx.JSON(http.StatusOK, Result{Msg: "登录成功"})
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十五分钟
			MaxAge: 30,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) edit(ctx *gin.Context) {
	type editReq struct {
		NickName string `json:"nickname"`
		BirthDay string `json:"birthday"`
		Brief    string `json:"aboutMe"`
	}
	req := editReq{}
	if err := ctx.Bind(&req); err != nil {
		return
	}

	nicknameLen := utf8.RuneCountInString(req.NickName)
	if nicknameLen > 30 || nicknameLen <= 0 {
		ctx.String(http.StatusOK, "昵称的长度在1-30之间")
		return
	}
	t, err := time.Parse(time.DateOnly, req.BirthDay)
	if err != nil {
		log.Printf("%v", err)
		ctx.String(http.StatusOK, "无法解析生日字符串")
		return
	}

	briefLen := utf8.RuneCountInString(req.Brief)
	if briefLen > 256 {
		ctx.String(http.StatusOK, "简介的长度在0-256之间")
		return
	}
	uc := ctx.MustGet("user").(UserClaims)
	id := uc.Uid
	err = h.svc.ModifyNoSensitiveInfo(ctx, &domain.User{Id: id,
		Nickname: req.NickName,
		AboutMe:  req.Brief,
		Birthday: t})
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, "修改成功")
}

func (h *UserHandler) profile(ctx *gin.Context) {
	us := ctx.MustGet("user").(UserClaims)

	id := us.Uid
	u, err := h.svc.Profile(ctx, &domain.User{Id: id})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	type UserInfo struct {
		NickName string `json:"Nickname"`
		BirthDay string `json:"Birthday"`
		Phone    string `json:"Phone"`
		Email    string `json:"Email"`
		AboutMe  string `json:"AboutMe"`
	}

	ctx.JSON(http.StatusOK, UserInfo{
		NickName: u.Nickname,
		BirthDay: u.Birthday.Format(time.DateOnly),
		AboutMe:  u.AboutMe,
		Email:    u.Email,
		Phone:    u.Phone,
	})
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统异常"})
		return
	}
	ok, err := h.phoneRexExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "手机格式不正确"})
		return
	}
	err = h.codeSvc.Send(ctx, bizLogin, req.Phone)

	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证失败"})
		return
	}

	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	// 设置登录态
	h.setJwtToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, Result{Msg: "登录成功"})
}
