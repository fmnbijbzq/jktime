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

var ErrDuplicateUser = repository.ErrDuplicateUser
var ErrInvalidUserOrPassword = errors.New("用户名或者密码不对")

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, u domain.User) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	Edit(ctx context.Context, u domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	bcryptd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("系统错误")

	}
	u.Password = string(bcryptd)
	err = svc.repo.Create(ctx, u)
	if err == ErrDuplicateUser {
		return err
	}
	return err
}

func (svc *userService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, u.Email)
	if err == repository.ErrUserNotFound {
		return user, ErrInvalidUserOrPassword
	}
	if err != nil {
		return user, errors.New("系统错误")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, err
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先找一下数据库，我们认为大部分用户都存在
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		// 有两种情况
		// 1. err == nil, u是可用的
		// 2. err != nil, 系统错误
		return u, err
	}
	// 用户没有找到
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	// 有两种错误用户, 唯一索引冲出（phone）
	// 一种是err != nil, 系统错误
	if err != nil && err != ErrDuplicateUser {
		return domain.User{}, err
	}
	// 要么err == nil, err==ErrDuplicateUser, 代表用户存在
	// 主从延迟，理论上来讲，强制走主库
	return svc.repo.FindByPhone(ctx, phone)

}

func (svc *userService) Edit(ctx context.Context, u domain.User) error {
	sess := sessions.Default(ctx.(*gin.Context))
	uid := sess.Get("userId")
	me, ok := uid.(int64)
	if !ok {
		return errors.New("系统出错")
	}
	u.Id = me
	return svc.repo.UpdateById(ctx, u)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, errors.New("系统出错")
	}
	return u, nil
}
