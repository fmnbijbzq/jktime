package login

import (
	"encoding/gob"
	"example/wb/internal/web"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginMiddlewareBuilder struct {
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
			path == "/user/login_sms" {
			return
		}
		authCode := ctx.Request.Header.Get("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]

		var uc web.UserClaims

		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(t *jwt.Token) (interface{}, error) {
			return web.JWTtoken, nil
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

		expireTime := uc.ExpiresAt
		if expireTime.Sub(time.Now()) < time.Second*50 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
			ss, err := token.SignedString(web.JWTtoken)
			if err != nil {
				log.Println(err)
			}
			ctx.Header("x-jwt-token", ss)
		}
		ctx.Set("user", uc)

	}
}
