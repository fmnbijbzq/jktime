package web

import (
	"example/wb/internal/domain"
	"example/wb/internal/service"
	"net/http"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	emailRegex = "^[a-z0-9A-Z]+[-|a-z0-9A-Z._]+@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-z]{2,}$"
	// 至少一个小写字母、至少一个大写字母、至少一个数字和至少一个特殊字符
	passwordRegex = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@!%*?&])[A-Za-z\d@!%*?&]{8,}$`
	birthdayRegex = `\d{4}-\d{1,2}-\d{1,2}`
)

var ErrDuplicateEmail = service.ErrDuplicateEmail

type UserHandler struct {
	EmailRegexExp    *regexp2.Regexp
	PasswordRegexExp *regexp2.Regexp
	BirthdayRegexExp *regexp2.Regexp
	svc              *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		EmailRegexExp:    regexp2.MustCompile(emailRegex, regexp2.None),
		PasswordRegexExp: regexp2.MustCompile(passwordRegex, regexp2.None),
		BirthdayRegexExp: regexp2.MustCompile(birthdayRegex, regexp2.None),
		svc:              svc,
	}
}

func (h *UserHandler) RegisterRoutes(g *gin.Engine) {
	ug := g.Group("/user")
	ug.POST("/signup", h.SignUp)
	ug.POST("/login", h.Login)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var sr SignUpReq
	if err := ctx.Bind(&sr); err != nil {
		return
	}
	isEmail, err := h.EmailRegexExp.MatchString(sr.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	if !isEmail {
		ctx.String(http.StatusOK, "email不合法")
	}
	if sr.Password != sr.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不同")
		return
	}
	isPasswd, err := h.PasswordRegexExp.MatchString(sr.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPasswd {
		ctx.String(http.StatusOK, "密码不合法")
		return
	}
	err = h.svc.SignUp(ctx.Request.Context(), domain.User{
		Email:    sr.Email,
		Password: sr.Password,
	})
	switch err {
	case ErrDuplicateEmail:
		ctx.String(http.StatusOK, ErrDuplicateEmail.Error())
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
	return
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Passowrd string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	u, err := h.svc.Login(ctx, domain.User{
		Email:    req.Email,
		Password: req.Passowrd,
	})
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			MaxAge: 600,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Profile(ctx *gin.Context) {

	sess := sessions.Default(ctx)
	uid := sess.Get("userId")
	me, ok := uid.(int64)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	u, err := h.svc.Profile(ctx, me)

	type Resp struct {
		NickName  string
		Birthday  string
		Biography string
	}
	switch err {
	case nil:
		ctx.JSONP(http.StatusOK, Resp{
			NickName:  u.NickName,
			Birthday:  u.Birthday.Format(time.DateOnly),
			Biography: u.Biography,
		})
	default:
		ctx.String(http.StatusOK, "系统出错")
	}
}

func (h *UserHandler) Edit(ctx *gin.Context) {

	type EditReq struct {
		NickName  string `json:"nickname"`
		Birthday  string `json:"birthday"`
		Biography string `json:"biography"`
	}

	var req EditReq

	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	const (
		nickLength      int = 50
		biographyLength int = 300
	)
	if len(req.NickName) >= nickLength {
		ctx.String(http.StatusOK, "昵称太长, 超过了50个字符")
		return
	}
	if len(req.NickName) == 0 {
		ctx.String(http.StatusOK, "昵称不能为空")
		return
	}
	if len(req.Biography) >= biographyLength {
		ctx.String(http.StatusOK, "个人简介的字数太长，超过了300个字符, 当前字符长度为", len(req.Biography))
		return
	}
	//
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	// ok, err := h.BirthdayRegexExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "日期格式不对, 请输入yyyy-mm-dd的格式")
		return
	}
	sess := sessions.Default(ctx)
	uid := sess.Get("userId")
	me, ok := uid.(int64)
	if !ok {
		ctx.String(http.StatusOK, "系统出错")
	}
	// t := carbon.Parse(req.Birthday).ToStdTime()
	err = h.svc.Edit(ctx, domain.User{
		Id:        me,
		NickName:  req.NickName,
		Birthday:  birthday,
		Biography: req.Biography,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "编辑成功")
	default:
		ctx.String(http.StatusOK, "系统出错")
	}
}
