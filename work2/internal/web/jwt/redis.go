package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var JWTtoken = []byte("mY2gT5iP0xZ9eX7tZ5eU9zI4lW0xP0wI")
var RCJWTkey = []byte("mY2gT5iP0xZ9eX7tZ5eU9zI4lixxP0wI")

type UserClaims struct {
	Id   int64
	Ssid string
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Uid  int64
	Ssid string
	jwt.RegisteredClaims
}

type RedisJWTHandler struct {
	client       redis.Cmdable
	JwtMethod    jwt.SigningMethod
	rcExpiration time.Duration
}

func NewJwtHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:       client,
		JwtMethod:    jwt.SigningMethodHS512,
		rcExpiration: time.Hour * 24 * 7,
	}
}

// CheckSession implements Handler.
func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.Request.Header.Get("Authorization")
	if authCode == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	tokenStr := segs[1]
	return tokenStr
}
func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return h.client.Set(ctx,
		fmt.Sprintf("users:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
}

func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}

	token := jwt.NewWithClaims(h.JwtMethod, claims)
	tokenStr, err := token.SignedString(RCJWTkey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {

	token := jwt.NewWithClaims(h.JwtMethod, UserClaims{
		Id:   uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间一分钟
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	})
	ss, err := token.SignedString(JWTtoken)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", ss)
	return nil
}
