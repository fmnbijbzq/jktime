package cache

import (
	"context"
	_ "embed"
	"errors"
	"example/wb/config"
	"example/wb/internal/repository/cache/redismock"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)

	}

	testCase := []struct {
		name string

		// redisMock func(ctrl *gomock.Controller) redis.Cmdable
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{ // TODO: Add test cases.
		{
			name: "测试成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				rcmd := redis.NewCmd(context.Background())
				rcmd.SetErr(nil)
				rcmd.SetVal(int64(0))

				cmd := redismock.NewMockCmdable(ctrl)
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode, []string{keyFunc("login_sms", "123456")}, "512364").
					Return(rcmd)

				return cmd
			},
			ctx:     context.Background(),
			biz:     "login_sms",
			phone:   "123456",
			code:    "512364",
			wantErr: nil,
		},
		{
			name: "redis返回error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				rcmd := redis.NewCmd(context.Background())
				rcmd.SetErr(errors.New("redis错误"))
				rcmd.SetVal(int64(0))

				cmd := redismock.NewMockCmdable(ctrl)
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode, []string{keyFunc("login_sms", "123456")}, "512364").
					Return(rcmd)

				return cmd
			},
			ctx:     context.Background(),
			biz:     "login_sms",
			phone:   "123456",
			code:    "512364",
			wantErr: errors.New("redis错误"),
		},
		{
			name: "验证码发送频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				rcmd := redis.NewCmd(context.Background())
				rcmd.SetErr(nil)
				rcmd.SetVal(int64(-1))

				cmd := redismock.NewMockCmdable(ctrl)
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode, []string{keyFunc("login_sms", "123456")}, "512364").
					Return(rcmd)

				return cmd
			},
			ctx:     context.Background(),
			biz:     "login_sms",
			phone:   "123456",
			code:    "512364",
			wantErr: ErrSendTooMany,
		},
		{
			name: "没有设置验证码时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				rcmd := redis.NewCmd(context.Background())
				rcmd.SetErr(nil)
				rcmd.SetVal(int64(-2))

				cmd := redismock.NewMockCmdable(ctrl)
				cmd.EXPECT().
					Eval(gomock.Any(), luaSetCode, []string{keyFunc("login_sms", "123456")}, "512364").
					Return(rcmd)

				return cmd
			},
			ctx:     context.Background(),
			biz:     "login_sms",
			phone:   "123456",
			code:    "512364",
			wantErr: errors.New("验证码存在，但没有过期时间"),
		},
	}
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisCmd := tt.mock(ctrl)
			redisCache := NewCodeCache(redisCmd)

			err := redisCache.Set(tt.ctx, tt.biz, tt.phone, tt.code)
			assert.Equal(t, tt.wantErr, err)
		})

	}
}

func TestRedisCodeCache_Set_e2e(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})

	testCase := []struct {
		name string

		before func(t *testing.T)

		after func(t *testing.T)

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",

			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				key := "phone_code:login:123456"
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9)
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "543210", code)
			},

			ctx:   context.Background(),
			biz:   "login",
			phone: "123456",
			code:  "543210",
		},
		{
			name: "系统错误",

			before: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				key := "phone_code:login:123456"
				err := rdb.Set(ctx, key, "543210", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				key := "phone_code:login:123456"
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur == -1)
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "543210", code)
			},

			ctx:   context.Background(),
			biz:   "login",
			phone: "123456",
			code:  "543210",

			wantErr: errors.New("验证码存在，但没有过期时间"),
		},
		{
			name: "验证码频繁发送",

			before: func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				key := "phone_code:login:123456"
				err := rdb.Set(ctx, key, "543210", time.Minute*10).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				key := "phone_code:login:123456"
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9)
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "543210", code)
			},

			ctx:   context.Background(),
			biz:   "login",
			phone: "123456",
			code:  "543210",

			wantErr: ErrSendTooMany,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(t)
			defer tt.after(t)

			c := NewCodeCache(rdb)
			err := c.Set(tt.ctx, tt.biz, tt.phone, tt.code)
			assert.Equal(t, tt.wantErr, err)

		})

	}
}
