package service

import (
	"context"
	"example/wb/internal/domain"
	"example/wb/internal/repository"
	"example/wb/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository
	l    logger.Logger
	// V1 写法专用, 两张不同的表再service层聚合
	// readerRepo repository.ArticleReaderRepository
	// authorRepo repository.ArticleAuthorRepository
}

// GetByAuthor implements ArticleService.
func (a *articleService) GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, limit, offset)

}

// Publish implements ArticleService.
func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

// Save implements ArticleService.
func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

// Withdraw implements ArticleService.
func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func NewArticleService(repo repository.ArticleRepository, l logger.Logger) ArticleService {
	return &articleService{
		repo: repo,
		l:    l,
	}
}
