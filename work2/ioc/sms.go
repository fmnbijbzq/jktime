package ioc

import (
	"example/wb/internal/repository"
	"example/wb/internal/service/sms"
	"example/wb/internal/service/sms/async"
	"example/wb/internal/service/sms/auth"
	"example/wb/internal/service/sms/failover"
	"example/wb/internal/service/sms/localsms"
	"example/wb/internal/service/sms/ratelimit"
	"example/wb/pkg/limiter"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitSMSService(redisCmd redis.Cmdable, repo repository.AsyncSmsRepository) sms.Service {
	sms1 := localsms.NewLocalService()
	sms2 := localsms.NewLocalService()
	ss := []sms.Service{sms1, sms2}
	fl := failover.NewTimeOutFailoverSMSService(ss, 10)

	au := auth.NewSMSService(fl)

	l := limiter.NewRedisLimter(redisCmd, time.Second, 1000)
	rls := ratelimit.NewRateLimitSMSService(au, l)
	return async.NewService(rls, repo)
	// return ratelimit.
}
