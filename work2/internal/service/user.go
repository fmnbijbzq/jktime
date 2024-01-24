package service

import (
	"context"
	"errors"
	"example/wb/internal/domain"
	"example/wb/internal/repository"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateEmail = repository.ErrDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("用户名或者密码不对")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	bcryptd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("系统错误")

	}
	u.Password = string(bcryptd)
	err = svc.repo.Create(ctx, u)
	if err == ErrDuplicateEmail {
		return err
	}
	return err
}

func (svc *UserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, u.Email)
	if err != nil {
		return user, ErrInvalidUserOrPassword
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		return user, ErrInvalidUserOrPassword
	}
	return user, err
}

func (svc *UserService) Edit(ctx context.Context, u domain.User) error {
	sess := sessions.Default(ctx.(*gin.Context))
	uid := sess.Get("userId")
	me, ok := uid.(int64)
	if !ok {
		return errors.New("系统出错")
	}
	u.Id = me
	return svc.repo.UpdateById(ctx, u)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, errors.New("系统出错")
	}
	return u, nil
}
