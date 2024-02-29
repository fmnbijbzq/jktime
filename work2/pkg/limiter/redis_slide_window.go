package limiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaScript string

type RedisLimter struct {
	cmd      redis.Cmdable
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisLimter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisLimter {
	return &RedisLimter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisLimter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaScript, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
