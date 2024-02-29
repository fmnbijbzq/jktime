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

func TestTimeOutFailoverSMSService_Send(t *testing.T) {
	tests := []struct {
		name string
		mock func(ctrl *gomock.Controller) []sms.Service

		idx       int32
		cnt       int32
		threshold int32

		wantIdx int32
		wantCnt int32
		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "不触发切换，成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmock.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				svc1 := smsmock.NewMockService(ctrl)
				return []sms.Service{svc0, svc1}
			},
			idx:       0,
			cnt:       8,
			threshold: 9,

			wantIdx: 0,
			wantCnt: 0,
			wantErr: nil,
		},
		{
			name: "触发切换，成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmock.NewMockService(ctrl)
				svc1 := smsmock.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				return []sms.Service{svc0, svc1}
			},
			idx:       0,
			cnt:       8,
			threshold: 5,

			wantIdx: 1,
			wantCnt: 0,
			wantErr: nil,
		},
		{
			name: "触发切换，触发超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmock.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.DeadlineExceeded)
				svc1 := smsmock.NewMockService(ctrl)
				return []sms.Service{svc0, svc1}
			},
			idx:       1,
			cnt:       8,
			threshold: 5,

			wantIdx: 0,
			wantCnt: 1,
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "触发切换，异常错误",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmock.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("0"))
				svc1 := smsmock.NewMockService(ctrl)
				return []sms.Service{svc0, svc1}
			},
			idx:       1,
			cnt:       8,
			threshold: 5,

			wantIdx: 0,
			wantCnt: 0,
			wantErr: errors.New("0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ss := tt.mock(ctrl)
			tfs := NewTimeOutFailoverSMSService(ss, tt.threshold)
			tfs.cnt = tt.cnt
			tfs.idx = tt.idx

			err := tfs.Send(context.Background(), "1", []string{"1"}, "123")

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantCnt, tfs.cnt)
			assert.Equal(t, tt.wantIdx, tfs.idx)

		})
	}
}
