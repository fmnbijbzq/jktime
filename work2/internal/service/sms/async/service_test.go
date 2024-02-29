package async

import (
	"context"
	"example/wb/internal/repository"
	repomock "example/wb/internal/repository/mock"
	"example/wb/internal/service/sms"
	smsmock "example/wb/internal/service/sms/mocks"
	"example/wb/internal/service/sms/ratelimit"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAsyncSMSService_Send(t *testing.T) {
	testCase := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository)
		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository) {
				smsSvc := smsmock.NewMockService(ctrl)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				repo := repomock.NewMockAsyncSmsRepository(ctrl)

				return smsSvc, repo
			},
			wantErr: nil,
		},
		{
			name: "触发限流存入数据库",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository) {
				smsSvc := smsmock.NewMockService(ctrl)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(ratelimit.ErrSMSLimitRate)

				repo := repomock.NewMockAsyncSmsRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil)
				return smsSvc, repo
			},
			wantErr: nil,
		},
	}
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			smsSvc, smsRepo := tt.mock(ctrl)
			asyncSvc := NewService(smsSvc, smsRepo)
			err := asyncSvc.Send(context.Background(), "1", []string{"12"}, "324")
			assert.Equal(t, tt.wantErr, err)

		})
	}
}

// func TestAsyncSMSService_AsyncSend(t *testing.T) {
// 	testCase := []struct {
// 		name    string
// 		mock    func(ctrl *gomock.Controller) (sms.Service, repository.SmsRepository, repository.CodeRepository)
// 		wantErr error
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			name: "发送成功",
// 			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SmsRepository, repository.CodeRepository) {
// 				smsSvc := smsmock.NewMockService(ctrl)
// 				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 					Return(nil)
// 				smsRepo := repomock.NewMockAsyncSmsRepository(ctrl)
// 				smsRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
// 					Return([]domain.Sms{
// 						{
// 							ID:    1,
// 							Phone: "11",
// 							Code:  "343253",
// 							TplId: "33",
// 						},
// 					}, nil)
// 				smsRepo.EXPECT().DeleteById(gomock.Any(), gomock.Any()).
// 					Return(nil)
// 				codeRepo := repomock.NewMockCodeRepository(ctrl)
// 				codeRepo.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
// 					Return(nil)

// 				return smsSvc, smsRepo, codeRepo
// 			},
// 			wantErr: nil,
// 		},
// 	}
// 	for _, tt := range testCase {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			smsSvc, smsRepo, codeRepo := tt.mock(ctrl)
// 			asyncSvc := NewAsyncService(smsSvc, smsRepo, codeRepo, time.Second, 3)
// 			err := asyncSvc.AsyncSend(context.Background(), "1")
// 			assert.Equal(t, tt.wantErr, err)
// 		})
// 	}
// }

func TestService_needAsync(t *testing.T) {
	testCase := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository)
		curTime   int
		threshold int
		durTimes  []int
		want      bool
		wantTimes []int
	}{
		// TODO: Add test cases.
		{
			name: "异步",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository) {
				svc := smsmock.NewMockService(ctrl)
				repo := repomock.NewMockAsyncSmsRepository(ctrl)
				return svc, repo
			},
			curTime:   600,
			threshold: 500,
			durTimes:  []int{400, 500},
			want:      true,
			wantTimes: []int{500, 600},
		},
		{
			name: "异步",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository) {
				svc := smsmock.NewMockService(ctrl)
				repo := repomock.NewMockAsyncSmsRepository(ctrl)
				return svc, repo
			},
			curTime:   600,
			threshold: 500,
			durTimes:  []int{400, 500, 500},
			want:      true,
			wantTimes: []int{500, 500, 600},
		},
		{
			name: "同步",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository) {
				svc := smsmock.NewMockService(ctrl)
				repo := repomock.NewMockAsyncSmsRepository(ctrl)
				return svc, repo
			},
			curTime:   400,
			threshold: 500,
			durTimes:  []int{400, 500},
			want:      false,
			wantTimes: []int{500, 400},
		},
	}
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo := tt.mock(ctrl)
			smsSvc := NewService(svc, repo)
			smsSvc.durTimes = tt.durTimes
			ok := smsSvc.needAsync(tt.curTime, tt.threshold)
			assert.Equal(t, tt.want, ok)
			assert.Equal(t, tt.wantTimes, smsSvc.durTimes)
		})
	}
}
