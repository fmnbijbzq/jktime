package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	ijwt "example/wb/internal/web/jwt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginMiddlewareBuilder(hdl ijwt.Handler) *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/signup" || path == "/user/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		const updateTimeKey = "update_time"
		gob.Register(time.Time{})

		val := sess.Get(updateTimeKey)
		t, ok := val.(time.Time)
		if val == nil || !ok || time.Now().Sub(t) >= time.Second*10 {
			sess.Set(updateTimeKey, time.Now())
			sess.Set("userId", userId)
			sess.Options(sessions.Options{
				Path:     "/user",
				MaxAge:   600,
				HttpOnly: true,
			})
			err := sess.Save()
			if err != nil {
				fmt.Println("session时间刷新失败")
			}
			return
		}
	}
}

func (m *LoginMiddlewareBuilder) CheckJWTLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/signup" ||
			path == "/user/login" ||
			path == "/user/hello" ||
			path == "/user/login_sms/code/send" ||
			path == "/user/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}
		tokenStr := m.ExtractToken(ctx)

		var uc ijwt.UserClaims

		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(t *jwt.Token) (interface{}, error) {
			return ijwt.JWTtoken, nil
		})
		if err != nil {
			// token 解析不出
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			// token 解析出来不对
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// token无效或者redis有问题
			// 过于严格
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 可以兼容redis异常时的情况
		// 同时要做好监控有没有error
		// 通常情况下使用这种写法
		// if cnt > 0 {
		// 	ctx.AbortWithStatus(http.StatusUnauthorized)
		// 	return
		// }

		ctx.Set("user", uc)

	}
}
