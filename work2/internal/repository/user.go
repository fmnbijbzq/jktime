package repository

import (
	"database/sql"
	"example/wb/internal/domain"
	"example/wb/internal/repository/cache"
	"example/wb/internal/repository/dao"
	"log"

	"golang.org/x/net/context"
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	UpdateById(ctx context.Context, u domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

var ErrDuplicateUser = dao.ErrDuplicateUser
var ErrUserNotFound = dao.ErrUserNotFound

func NewCachedUserRepository(dao dao.UserDao, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	err := repo.dao.Insert(ctx, repo.toEntity(u))
	if err == ErrDuplicateUser {
		return ErrDuplicateUser

	}
	return err
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil

}

func (repo *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	uc, err := repo.cache.Get(ctx, id)
	if err == nil {
		return uc, err
	}
	// redis 中err有两种可能：
	// 1. redis崩溃了，网络出错了
	// 2. redis是正常的，redis中不存在当前key
	u, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	var du = repo.toDomain(u)
	go func() {
		err = repo.cache.Set(ctx, du)
		if err != nil {
			log.Println(err)
		}
	}()
	return du, nil
}

func (repo *CachedUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:        u.ID,
		NickName:  u.NickName,
		Birthday:  u.Birthday,
		Biography: u.Biography,
		Email:     u.Email.String,
		Phone:     u.Phone.String,
		Password:  u.Password,
	}

}

func (repo *CachedUserRepository) UpdateById(ctx context.Context, u domain.User) error {

	return repo.dao.UpdateById(ctx, dao.User{
		ID:        u.Id,
		NickName:  u.NickName,
		Birthday:  u.Birthday,
		Biography: u.Biography,
	})

}

func (repo *CachedUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		ID: u.Id,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password:  u.Password,
		Birthday:  u.Birthday,
		Biography: u.Biography,
		NickName:  u.NickName,
	}

}
