package service_test

import (
	"context"
	"errors"
	"example/wb/internal/domain"
	"example/wb/internal/repository"
	repomock "example/wb/internal/repository/mock"
	"example/wb/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("dioa@W524")

	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))

}

func TestUserService_Login(t *testing.T) {
	testCase := []struct {
		name string

		mock func(ctrl *gomock.Controller) repository.UserRepository

		u domain.User

		wantErr error

		wantUser domain.User
	}{
		{
			name: "登录成功",

			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "12321@qq.com").
					Return(domain.User{
						Email:    "12321@qq.com",
						Password: "$2a$10$wkH/UH/yQix09TOq.CIE3O3b.m84KuROb3/E90GU9QYHistAjbTCm",
					}, nil)
				return repo
			},

			u: domain.User{
				Email:    "12321@qq.com",
				Password: "dioa@W524",
			},

			wantErr: nil,

			wantUser: domain.User{
				Email:    "12321@qq.com",
				Password: "$2a$10$wkH/UH/yQix09TOq.CIE3O3b.m84KuROb3/E90GU9QYHistAjbTCm",
			},
		},
		{
			name: "系统错误",

			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "12321@qq.com").
					Return(domain.User{}, errors.New("db错误"))
				return repo
			},

			u: domain.User{
				Email:    "12321@qq.com",
				Password: "dioa@W524",
			},

			wantErr: errors.New("系统错误"),

			wantUser: domain.User{},
		},
		{
			name: "密码不对",

			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomock.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "12321@qq.com").
					Return(domain.User{
						Email:    "12321@qq.com",
						Password: "$2a$10$wkH/UH/yQix09TOq.CIE3O3b.m84KuROb3/E90GU9QYHistAjbTCm",
					}, nil)
				return repo
			},

			u: domain.User{
				Email:    "12321@qq.com",
				Password: "dioa@W52334",
			},

			wantErr: service.ErrInvalidUserOrPassword,

			wantUser: domain.User{},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := service.NewUserService(repo)

			var c context.Context
			u, err := svc.Login(c, tc.u)

			assert.Equal(t, err, tc.wantErr)
			assert.Equal(t, u, tc.wantUser)

		})

	}

}
