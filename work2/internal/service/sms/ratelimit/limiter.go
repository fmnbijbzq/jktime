package ratelimit

import (
	"context"
	"errors"
	"example/wb/internal/service/sms"
	"example/wb/pkg/limiter"
)

var ErrSMSLimitRate = errors.New("短信发送频繁，触发了限流")

// 此种情况的svc需要实现sms.Service
// 接口中的所有方法
// 好处是用户不能绕开你的装饰器
type RateLimitSMSService struct {
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

// 使用组合的形式创建装饰器
// 用户可以通过字段拿到接口中其他参数
// 如果接口中有多个参数，可以仅实现部分参数
// type RateLimitSMSServiceV1 struct {
// 	sms.Service
// 	limiter limiter.Limiter
// 	key     string
// }

func NewRateLimitSMSService(svc sms.Service, l limiter.Limiter) sms.Service {

	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms_ratelimit",
	}

}

func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return ErrSMSLimitRate
	}
	if limited {
		return ErrSMSLimitRate

	}
	return r.svc.Send(ctx, tplId, args, numbers...)
}
