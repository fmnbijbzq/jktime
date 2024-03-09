package repository

import (
	"context"
	"example/wb/internal/domain"
	"example/wb/internal/repository/cache"
	"example/wb/internal/repository/dao"

	"github.com/ecodeclub/ekit/slice"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache
}

// GetByAuthor implements ArticleRepository.
func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, error) {
	// 事实上，limit <= 100都可以走缓存
	// 但是要注意需要自己通过limit控制返回的数据
	if offset == 0 && limit == 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中,你是可以忽略的
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return []domain.Article{}, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return toDomain(src)
	})
	go func() {
		if offset == 0 && limit == 100 {
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()
	return res, nil

}

// Create implements ArticleRepository.
func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, toEntity(art))

	if err == nil {
		err = c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			// 记录日志

		}
	}

	return id, err
}

// Sync implements ArticleRepository.
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, toEntity(art))
	if err == nil {
		err = c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			// 记录日志

		}
	}
	return id, err
}

// SyncStatus implements ArticleRepository.
func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, uint8(status))
	if err == nil {
		err = c.cache.DelFirstPage(ctx, uid)
		if err != nil {
			// 记录日志

		}
	}
	return err
}

// Update implements ArticleRepository.
func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, toEntity(art))

	if err == nil {
		err = c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			// 记录日志

		}
	}
	return err
}

func toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	}

}

func toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
	}

}

func NewArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		cache: cache,
	}
}
