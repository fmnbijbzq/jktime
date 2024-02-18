package repository

import (
	"context"
	"example/wb/internal/repository/cache"
)

var ErrSendTooMany = cache.ErrSendTooMany
var ErrCodeVertifyTooMany = cache.ErrCodeVertifyTooMany

type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error
	Vertify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCachedCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{cache: cache}
}

func (c *CachedCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return c.cache.Set(ctx, biz, phone, code)
}

func (c *CachedCodeRepository) Vertify(ctx context.Context, biz, phone, code string) (bool, error) {
	return c.cache.Vertify(ctx, biz, phone, code)
}
