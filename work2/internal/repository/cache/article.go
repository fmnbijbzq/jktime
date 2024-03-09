package cache

import (
	"context"
	"encoding/json"
	"example/wb/internal/domain"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

// DelFirstPage implements ArticleCache.
func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	return a.client.Del(ctx, a.key(uid)).Err()
}

// GetFirstPage implements ArticleCache.
func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.key(uid)
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

// SetFirstPage implements ArticleCache.
func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstact()
	}
	key := a.key(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) key(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{
		client: client,
	}

}
