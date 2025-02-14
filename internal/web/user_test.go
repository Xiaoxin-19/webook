package web

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webok/internal/domain"
	"webok/internal/service"
	svcmocks "webok/internal/service/mock"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/logger"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler)
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		wantBody   string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), &domain.User{
					Email:    "123@qq.com",
					Password: "pass@1234",
				}).Return(nil)

				return userSvc, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123@qq.com",
							"password": "pass@1234",
							"confirmPassword": "pass@1234"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"注册成功","data":null}`,
		},
		{
			name: "Bind绑定失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123@qq.com",
							"password": "pass@1234",
							"confirmPassword": "pass@1234"
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123q.com",
							"password": "pass@1234",
							"confirmPassword": "pass@1234"
							}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"邮箱格式错误","data":null}`,
		},
		{
			name: "两次密码输入不同",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123@qq.com",
							"password": "pass@1234",
							"confirmPassword": "pass@123"
							}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"两次密码不一致","data":null}`,
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123@qq.com",
							"password": "123456",
							"confirmPassword": "123456"
							}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"密码格式错误","data":null}`,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), &domain.User{
					Email:    "123@qq.com",
					Password: "pass@1234",
				}).Return(errors.New("error in db"))
				return userSvc, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123@qq.com",
							"password": "pass@1234",
							"confirmPassword": "pass@1234"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), &domain.User{
					Email:    "123@qq.com",
					Password: "pass@1234",
				}).Return(service.ErrDuplicate)
				return userSvc, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`
							{
							"email": "123@qq.com",
							"password": "pass@1234",
							"confirmPassword": "pass@1234"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"邮箱冲突","data":null}`,
		},
	}
	var l logger.Logger = logger.NewNopLogger()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 不要全部测试用例共用一个ctrl
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userSvc, codeSvc, jwtHdl := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc, jwtHdl, l)

			server := gin.Default()
			hdl.RegisterRoutes(server)

			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}
