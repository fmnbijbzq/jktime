package auth

import (
	"context"
	"example/wb/internal/service/sms"

	"github.com/golang-jwt/jwt/v5"
)

type SMSSerivce struct {
	svc sms.Service
	key []byte
}

func NewSMSService(svc sms.Service) *SMSSerivce {
	return &SMSSerivce{
		svc: svc,
		key: []byte("mY2gT5iP0xZ9eX7tZ5eU9zIdfl0xP0wI"),
	}

}

func (s *SMSSerivce) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims UserClaims
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(t *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, claims.tplId, args, numbers...)
}

type UserClaims struct {
	jwt.RegisteredClaims
	tplId string
}
