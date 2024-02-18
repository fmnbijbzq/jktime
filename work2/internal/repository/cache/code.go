package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/vertify_code.lua
	luaVertifyCode string

	ErrSendTooMany        = errors.New("验证码发送太频繁")
	ErrCodeVertifyTooMany = errors.New("验证码验证太频繁")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Vertify(ctx context.Context, biz, phone, code string) (bool, error)
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{cmd: cmd}
}

func (cache *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cache.cmd.Eval(ctx, luaSetCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case -2:
		return errors.New("验证码存在，但没有过期时间")
	case -1:
		return ErrSendTooMany
	default:
		return nil

	}
}

func (cache *RedisCodeCache) Vertify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := cache.cmd.Eval(ctx, luaVertifyCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVertifyTooMany
	default:
		return true, nil

	}
}

func (cache *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type CodeLocalCache struct {
	mu    sync.Mutex
	cache *freecache.Cache
}

func NewCodeLocalCache(cache *freecache.Cache) CodeCache {
	return &CodeLocalCache{
		cache: cache,
	}
}

func (c *CodeLocalCache) Set(ctx context.Context, biz, phone, code string) error {
	curKey := c.key(biz, phone)
	cntKey := curKey + ":cnt"
	c.mu.Lock()
	defer c.mu.Unlock()
	lfTime, err := c.cache.TTL([]byte(curKey))
	// key存在且验证码的存在时间在60秒以内
	if err == nil && lfTime > 540 {
		return ErrSendTooMany
	}
	// key不存在或者验证码存在时间超过了60秒, 重新设置验证码
	err = c.cache.Set([]byte(curKey), []byte(code), 600)
	if err != nil {
		return errors.New("系统错误")
	}
	err = c.cache.Set([]byte(cntKey), []byte{3}, 600)
	if err != nil {
		return errors.New("系统错误")
	}
	return nil
}

func (c *CodeLocalCache) Vertify(ctx context.Context, biz, phone, code string) (bool, error) {
	curKey := c.key(biz, phone)
	cntKey := curKey + ":cnt"
	c.mu.Lock()
	defer c.mu.Unlock()
	val, err := c.cache.Get([]byte(cntKey))
	count := val[0]
	if err != nil || count <= 0 {
		return false, err
	}
	if count <= 0 {
		return false, ErrCodeVertifyTooMany
	}
	val, err = c.cache.Get([]byte(curKey))
	if err != nil {
		return false, err
	}

	if string(val) == code {
		return true, nil
	} else {
		count = count - 1
		c.cache.Set([]byte(cntKey), []byte{count}, 600)
		return false, nil

	}
}

func (cache *CodeLocalCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
