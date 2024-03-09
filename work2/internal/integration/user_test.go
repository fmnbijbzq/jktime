package integration

import (
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// func TestUserHandler_SendSMSCode(t *testing.T) {
// 	rdb := startup.InitRedis()
// 	server := startup.InitWebServer()

// 	testCases := []struct {
// 		name string

// 		// 准备数据
// 		before func(t *testing.T)

// 		//验证数据
// 		after func(t *testing.T)

// 		phone string

// 		wantCode int
// 		wantBody web.Result
// 	}{
// 		{
// 			name: "未输入手机号码",

// 			before: func(t *testing.T) {
// 			},
// 			after: func(t *testing.T) {
// 			},
// 			phone:    "",
// 			wantCode: http.StatusOK,
// 			wantBody: web.Result{
// 				Code: 4,
// 				Msg:  "请输入手机号码",
// 			},
// 		},
// 		{
// 			name: "发送成功用例",

// 			before: func(t *testing.T) {

// 			},
// 			after: func(t *testing.T) {
// 				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 				defer cancel()

// 				key := "phone_code:login:1234567"
// 				code, err := rdb.Get(ctx, key).Result()
// 				assert.NoError(t, err)
// 				assert.True(t, len(code) == 6)
// 				dur, err := rdb.TTL(ctx, key).Result()
// 				assert.NoError(t, err)
// 				assert.True(t, dur > time.Minute*9)
// 				err = rdb.Del(ctx, key).Err()
// 				assert.NoError(t, err)
// 			},
// 			phone:    "1234567",
// 			wantCode: http.StatusOK,
// 			wantBody: web.Result{
// 				Msg: "发送成功",
// 			},
// 		},
// 		{
// 			name: "验证码发送太频繁",

// 			before: func(t *testing.T) {
// 				ctx, cancel := context.WithCancel(context.Background())
// 				defer cancel()

// 				key := "phone_code:login:1234567"
// 				err := rdb.Set(ctx, key, "123456", time.Minute*10).Err()
// 				assert.NoError(t, err)

// 			},
// 			after: func(t *testing.T) {
// 				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 				defer cancel()

// 				key := "phone_code:login:1234567"
// 				code, err := rdb.GetDel(ctx, key).Result()
// 				assert.NoError(t, err)
// 				assert.Equal(t, "123456", code)
// 			},
// 			phone:    "1234567",
// 			wantCode: http.StatusOK,
// 			wantBody: web.Result{
// 				Code: 4,
// 				Msg:  "短信发送太频繁，请稍后再试",
// 			},
// 		},
// 		{
// 			name: "系统错误",

// 			before: func(t *testing.T) {
// 				ctx, cancel := context.WithCancel(context.Background())
// 				defer cancel()

// 				key := "phone_code:login:1234567"
// 				err := rdb.Set(ctx, key, "123456", 0).Err()
// 				assert.NoError(t, err)

// 			},
// 			after: func(t *testing.T) {
// 				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 				defer cancel()

// 				key := "phone_code:login:1234567"
// 				code, err := rdb.GetDel(ctx, key).Result()
// 				assert.NoError(t, err)
// 				assert.Equal(t, "123456", code)
// 			},
// 			phone:    "1234567",
// 			wantCode: http.StatusOK,
// 			wantBody: web.Result{
// 				Code: 5,
// 				Msg:  "系统错误",
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.before(t)
// 			defer tc.after(t)

// 			// 注册路由
// 			req, err := http.NewRequest(http.MethodPost,
// 				"/user/login_sms/code/send",
// 				bytes.NewReader([]byte(fmt.Sprintf(`{"phone":"%s"}`, tc.phone))))
// 			assert.NoError(t, err)
// 			req.Header.Set("Content-Type", "application/json")
// 			recorder := httptest.NewRecorder()

// 			// 执行
// 			server.ServeHTTP(recorder, req)

// 			// 执行结果
// 			assert.Equal(t, tc.wantCode, recorder.Code)
// 			var res web.Result
// 			err = json.NewDecoder(recorder.Body).Decode(&res)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tc.wantBody, res)

// 		})
// 	}
// }
