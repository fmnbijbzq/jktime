package ratelimit

import (
	"context"
	"errors"
	"example/wb/internal/service/sms"
	smsmock "example/wb/internal/service/sms/mocks"
	"example/wb/pkg/limiter"
	limitmock "example/wb/pkg/limiter/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRateLimitSMSService_Send(t *testing.T) {
	testCase := []struct {
		name string

		mock func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter)

		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				sms := smsmock.NewMockService(ctrl)
				sms.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				limit := limitmock.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return sms, limit
			},
			wantErr: nil,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				sms := smsmock.NewMockService(ctrl)
				// sms.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				limit := limitmock.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return sms, limit
			},
			wantErr: ErrSMSLimitRate,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				sms := smsmock.NewMockService(ctrl)
				// sms.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				limit := limitmock.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("系统错误"))
				return sms, limit
			},
			wantErr: errors.New("系统错误"),
		},
	}
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sms, limit := tt.mock(ctrl)
			rl := NewRateLimitSMSService(sms, limit)
			err := rl.Send(context.Background(), "1", []string{"1"}, "213123")
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
