package web

import (
	"example/wb/internal/domain"
	"example/wb/internal/service"
	"net/http"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	emailRegex = "^[a-z0-9A-Z]+[-|a-z0-9A-Z._]+@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-z]{2,}$"
	// 至少一个小写字母、至少一个大写字母、至少一个数字和至少一个特殊字符
	passwordRegex = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@!%*?&])[A-Za-z\d@!%*?&]{8,}$`
	birthdayRegex = `\d{4}-\d{1,2}-\d{1,2}`
	bizLogin      = "login"
)

var ErrSendTooMany = service.ErrSendTooMany
var ErrCodeVertifyTooMany = service.ErrCodeVertifyTooMany
var ErrDuplicateUser = service.ErrDuplicateUser

type UserHandler struct {
	EmailRegexExp    *regexp2.Regexp
	PasswordRegexExp *regexp2.Regexp
	BirthdayRegexExp *regexp2.Regexp
	svc              service.UserService
	codeSvc          service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		EmailRegexExp:    regexp2.MustCompile(emailRegex, regexp2.None),
		PasswordRegexExp: regexp2.MustCompile(passwordRegex, regexp2.None),
		BirthdayRegexExp: regexp2.MustCompile(birthdayRegex, regexp2.None),
		svc:              svc,
		codeSvc:          codeSvc,
	}
}

func (h *UserHandler) RegisterRoutes(g *gin.Engine) {
	ug := g.Group("/user")
	ug.GET("/hello", h.Hello)
	ug.POST("/signup", h.SignUp)
	// ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) Hello(ctx *gin.Context) {
	ctx.String(http.StatusOK, "hello")
	return

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
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "email不合法")
		return
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
	case ErrDuplicateUser:
		ctx.String(http.StatusOK, ErrDuplicateUser.Error())
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

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
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
		// token := jwt.NewWithClaims(jwt.SigningMethodHS512, UserClaims{
		// 	Id: u.Id,
		// 	RegisteredClaims: jwt.RegisteredClaims{
		// 		// 过期时间一分钟
		// 		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		// 		NotBefore: jwt.NewNumericDate(time.Now()),
		// 	},
		// })
		// ss, err := token.SignedString(JWTtoken)
		// if err != nil {
		// 	ctx.String(http.StatusOK, "系统错误")
		// }
		// ctx.Header("x-jwt-token", ss)
		h.setJWTToken(ctx, u.Id)
		// sess := sessions.Default(ctx)
		// sess.Set("userId", u.Id)
		// sess.Options(sessions.Options{
		// 	MaxAge: 600,
		// })
		// err = sess.Save()
		// if err != nil {
		// 	ctx.String(http.StatusOK, "系统错误")
		// 	return
		// }
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) setJWTToken(ctx *gin.Context, uid int64) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, UserClaims{
		Id: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间一分钟
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	})
	ss, err := token.SignedString(JWTtoken)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.Header("x-jwt-token", ss)
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		// 会帮助我们写回一个400的错误
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
		return
	}

	err := h.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case ErrSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 补日志
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
	ok, err := h.codeSvc.Vertify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		})
		return
	}
	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	h.setJWTToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})

}

func (h *UserHandler) Profile(ctx *gin.Context) {

	uc, exists := ctx.Get("user")
	if !exists {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	me, ok := uc.(UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	u, err := h.svc.Profile(ctx, me.Id)

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
	uc, exists := ctx.Get("user")
	if !exists {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	me, ok := uc.(UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	// sess := sessions.Default(ctx)
	// uid := sess.Get("userId")
	// me, ok := uid.(int64)
	// if !ok {
	// 	ctx.String(http.StatusOK, "系统出错")
	// }
	// t := carbon.Parse(req.Birthday).ToStdTime()
	err = h.svc.Edit(ctx, domain.User{
		Id:        me.Id,
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

var JWTtoken = []byte("mY2gT5iP0xZ9eX7tZ5eU9zI4lW0xP0wI")

type UserClaims struct {
	Id int64
	jwt.RegisteredClaims
}
