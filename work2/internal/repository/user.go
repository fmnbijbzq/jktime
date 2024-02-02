package repository

import (
	"example/wb/internal/domain"
	"example/wb/internal/repository/dao"

	"golang.org/x/net/context"
)

type UserRepository struct {
	dao *dao.UserDao
}

var ErrDuplicateEmail = dao.ErrDuplicateEmail

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	err := repo.dao.Insert(ctx, repo.toEntity(u))
	if err == ErrDuplicateEmail {
		return ErrDuplicateEmail

	}
	return err
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{Id: u.ID, Email: u.Email, Password: u.Password}, nil
}
func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:        u.ID,
		NickName:  u.NickName,
		Birthday:  u.Birthday,
		Biography: u.Biography,
		Email:     u.Email,
		Password:  u.Password,
	}, nil
}

func (repo *UserRepository) UpdateById(ctx context.Context, u domain.User) error {

	return repo.dao.UpdateById(ctx, dao.User{
		ID:        u.Id,
		NickName:  u.NickName,
		Birthday:  u.Birthday,
		Biography: u.Biography,
	})

}

func (repo *UserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Email:    u.Email,
		Password: u.Password,
	}

}
