package web_test

import (
	"bytes"
	"example/wb/internal/domain"
	"example/wb/internal/service"
	svcmock "example/wb/internal/service/mocks"
	"example/wb/internal/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	// "github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		// 准备mock
		// 因为UserHandler 用到了UserService和CodeService
		// 所以我们需要准备两个mock实例
		// 因此你能看到它返回了UserService和CodeService
		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

		// 输入, 因为request的构造过程很复杂
		// 所以我们定义一个builder
		reqBuilder func(t *testing.T) *http.Request

		// 预取响应
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmock.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "14526@qq.com",
					Password: "helLo@dsf46",
				}).Return(nil)
				codeSvc := svcmock.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewReader([]byte(`
				{
					"email": "14526@qq.com",
					"password": "helLo@dsf46",
					"confirmPassword": "helLo@dsf46"
				}
				`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req

			},
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "Bind出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmock.NewMockUserService(ctrl)
				// userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
				// 	Email:    "14526@qq.com",
				// 	Password: "helLo@dsf46",
				// }).Return(nil)
				codeSvc := svcmock.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewReader([]byte(`
				{
					"email": "14526@qq.com",
					"password": "helLo@dsf46"
					"confirmPassword": "helLo@dsf46"
				}
				`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req

			},
			wantCode: http.StatusBadRequest,
			wantBody: "",
		},
		{
			name: "邮箱出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmock.NewMockUserService(ctrl)
				// userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
				// 	Email:    "14526qq.com",
				// 	Password: "helLo@dsf46",
				// }).Return(nil)
				codeSvc := svcmock.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewReader([]byte(`
				{
					"email": "14526qq.com",
					"password": "helLo@dsf46",
					"confirmPassword": "helLo@dsf46"
				}
				`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req

			},
			wantCode: http.StatusOK,
			wantBody: "email不合法",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmock.NewMockUserService(ctrl)
				// userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
				// 	Email:    "14526@qq.com",
				// 	Password: "helLo@dsf46",
				// }).Return(nil)
				codeSvc := svcmock.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewReader([]byte(`
				{
					"email": "14526@qq.com",
					"password": "helLodsf46",
					"confirmPassword": "helLodsf46"
				}
				`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req

			},
			wantCode: http.StatusOK,
			wantBody: "密码不合法",
		},
		{
			name: "两次密码不同",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmock.NewMockUserService(ctrl)
				// userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
				// 	Email:    "14526@qq.com",
				// 	Password: "helLo@dsf46",
				// }).Return(nil)
				codeSvc := svcmock.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewReader([]byte(`
				{
					"email": "14526@qq.com",
					"password": "helLo@dsf46",
					"confirmPassword": "helLo@sf46"
				}
				`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req

			},
			wantCode: http.StatusOK,
			wantBody: "两次密码不同",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 利用mock生成Service
			userSvc, codeSvc := tc.mock(ctrl)
			hdl := web.NewUserHandler(userSvc, codeSvc)

			// 注册路由
			server := gin.Default()
			hdl.RegisterRoutes(server)

			// 准备req和记录的 recoder
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()

			// 执行
			server.ServeHTTP(recorder, req)

			// 执行结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			t.Log(recorder.Body.String())
			assert.Equal(t, tc.wantBody, recorder.Body.String())

		})
	}
}

// func TestHandler_S(t *testing.T) {
// 	testCase := []struct {
// 		name string

// 		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

// 		reqBuilder func(t *testing.T) *http.Request

// 		wantCode int
// 		wantBody string
// 	}{
// 		{},
// 	}

// 	for _, tc := range testCase {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			userSvc, codeSvc := tc.mock(ctrl)
// 			hdl := web.NewUserHandler(userSvc, codeSvc)

// 			server := gin.Default()
// 			hdl.RegisterRoutes(server)

// 			req := tc.reqBuilder(t)
// 			recoverd := httptest.NewRecorder()

// 			server.ServeHTTP(recoverd, req)

// 			assert.Equal(t, tc.wantCode, recoverd.Code)
// 			assert.Equal(t, tc.wantBody, recoverd.Body.String())

// 		},
// 		)

// 	}
// }
