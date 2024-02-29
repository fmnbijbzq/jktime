package failover

import (
	"context"
	"errors"
	"example/wb/internal/service/sms"
	smsmock "example/wb/internal/service/sms/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFailOverSMSService_Send(t *testing.T) {
	testCase := []struct {
		name string

		mock func(ctrl *gomock.Controller) []sms.Service

		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "二次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				var svcs []sms.Service
				for {
					svc := smsmock.NewMockService(ctrl)
					svcs = append(svcs, svc)
					if len(svcs) == 2 {
						svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(nil)
						return svcs
					} else {
						svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(errors.New("er"))
					}
				}
			},
			wantErr: nil,
		},
		{
			name: "轮询失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				var svcs []sms.Service
				for {
					svc := smsmock.NewMockService(ctrl)
					svcs = append(svcs, svc)
					if len(svcs) == 2 {
						svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(errors.New("轮询了所有服务商, 但是都失败了"))
						return svcs
					} else {
						svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(errors.New("er"))
					}
				}
			},
			wantErr: errors.New("轮询了所有服务商, 但是都失败了"),
		},
	}
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svcs := tt.mock(ctrl)
			fo := NewFailOverSMSService(svcs)
			err := fo.Send(context.Background(), "1", []string{"1"}, "213")
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
