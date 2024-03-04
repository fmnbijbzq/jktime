package web

import (
	"example/wb/internal/service"
	"example/wb/internal/service/oauth2/wechat"
	ijwt "example/wb/internal/web/jwt"
	"example/wb/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
)

type OAuth2WechatHandler struct {
	ijwt.Handler
	svc             wechat.Service
	userSvc         service.UserService
	key             []byte
	cookieStateName string
	l               logger.Logger
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.Auth2Url)
	g.Any("/callback", o.CallBack)

}

func NewOAuth2WechatHandler(svc wechat.Service,
	hdl ijwt.Handler,
	l logger.Logger,
	userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("mY2gT5iP0xZ9eX7tZ5eU9zI4no0xP0wI"),
		cookieStateName: "jwt-state",
		Handler:         hdl,
		l:               l,
	}

}

func (o *OAuth2WechatHandler) Auth2Url(ctx *gin.Context) {
	state := uuid.New()
	val, err := o.svc.AUthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造跳转URL失败",
		})
	}
	o.setStateCookie(ctx, state)
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})
	// ctx.Redirect(http.StatusFound, val)
}

func (o *OAuth2WechatHandler) CallBack(ctx *gin.Context) {
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "非法请求",
		})
		return
	}
	// 校验不校验都可以
	code := ctx.Query("code")
	// state := ctx.Query("state")
	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "授权码有误",
		})
		return
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = o.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")

	ck, err := ctx.Cookie(h.cookieStateName)
	if err != nil {
		return fmt.Errorf("无法获得cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(t *jwt.Token) (interface{}, error) {
		return h.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析token失败 %w", err)
	}
	if state != sc.state {
		// state不匹配
		return fmt.Errorf("state不匹配")
	}
	return nil
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {

	claims := StateClaims{
		state: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.key)
	if err != nil {
		return err
	}
	// secure 是否使用https
	ctx.SetCookie(h.cookieStateName, tokenStr, 600,
		"/oauth2/wechat/callback", "", false, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	state string
}
