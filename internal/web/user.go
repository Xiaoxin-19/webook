package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
)

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", h.signUp)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.edit)
	ug.GET("/profile", h.profile)
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
	case errors.Is(err, service.ErrDuplicateEmail):
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
	switch err {
	case nil:
		uc := UserClaims{
			Uid:       u.Id,
			UserAgent: ctx.GetHeader("User-Agent"),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		tokenStr, err := token.SignedString(JWTKey)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
		}
		ctx.Header("x-jwt-token", tokenStr)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
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
	err = h.svc.ModifyNoSensitiveInfo(ctx, &domain.User{Id: id, Nickname: req.NickName, Brief: req.Brief, Birthday: t.UnixMilli()})
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
		BirthDay int64  `json:"Birthday"`
		Phone    string `json:"Phone"`
		Email    string `json:"Email"`
		AboutMe  string `json:"AboutMe"`
	}

	ctx.JSONP(http.StatusOK, UserInfo{
		NickName: u.Nickname,
		BirthDay: u.Birthday,
		AboutMe:  u.Brief,
		Email:    u.Email,
	})
}
