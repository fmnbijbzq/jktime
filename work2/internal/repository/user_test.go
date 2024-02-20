package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"example/wb/internal/domain"
	"example/wb/internal/repository"
	"example/wb/internal/repository/cache"
	cachemock "example/wb/internal/repository/cache/mock"
	"example/wb/internal/repository/dao"
	daomock "example/wb/internal/repository/dao/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCase := []struct {
		name string

		mock func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache)

		id      int64
		want    domain.User
		wantErr error
	}{
		{
			name: "db查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				userDao := daomock.NewMockUserDao(ctrl)
				userDao.EXPECT().
					FindById(gomock.Any(), int64(0)).
					Return(dao.User{
						Email: sql.NullString{
							String: "test@qq.com",
							Valid:  true,
						},
						Password:  "111",
						Biography: "324dfsoanj",
					}, nil)

				userCache := cachemock.NewMockUserCache(ctrl)
				userCache.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(domain.User{}, errors.New("Not Found"))
				return userDao, userCache
			},
			id: 0,
			want: domain.User{
				Email:     "test@qq.com",
				Password:  "111",
				Biography: "324dfsoanj",
			},
			wantErr: nil,
		},
		{
			name: "缓存查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				userDao := daomock.NewMockUserDao(ctrl)
				// userDao.EXPECT().
				// 	FindById(gomock.Any(), int64(0)).
				// 	Return(dao.User{
				// 		Email: sql.NullString{
				// 			String: "test@qq.com",
				// 			Valid:  true,
				// 		},
				// 		Password:  "111",
				// 		Biography: "324dfsoanj",
				// 	}, nil)

				userCache := cachemock.NewMockUserCache(ctrl)
				userCache.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(domain.User{
						Email:     "test@qq.com",
						Password:  "111",
						Biography: "324dfsoanj",
					}, nil)
				return userDao, userCache
			},
			id: 0,
			want: domain.User{
				Email:     "test@qq.com",
				Password:  "111",
				Biography: "324dfsoanj",
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，数据库不存在",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				userDao := daomock.NewMockUserDao(ctrl)
				userDao.EXPECT().
					FindById(gomock.Any(), int64(0)).
					Return(dao.User{}, dao.ErrUserNotFound)

				userCache := cachemock.NewMockUserCache(ctrl)
				userCache.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(domain.User{}, errors.New("Not Found"))
				return userDao, userCache
			},
			id:      0,
			want:    domain.User{},
			wantErr: dao.ErrUserNotFound,
		},
		// TODO: Add test cases.
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			userDao, userCache := tc.mock(ctrl)
			userRepo := repository.NewCachedUserRepository(userDao, userCache)

			var ctx context.Context
			u, err := userRepo.FindById(ctx, tc.id)

			assert.Equal(t, err, tc.wantErr)
			assert.Equal(t, u, tc.want)

		})
	}
}
