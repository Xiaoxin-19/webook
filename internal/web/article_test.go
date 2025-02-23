package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webok/internal/service"
	svcmocks "webok/internal/service/mock"
	ijwt "webok/internal/web/jwt"
	"webok/pkg/logger"
)

type Result[T any] struct {
	Code uint16 `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func TestArticleHandler_publish(t *testing.T) {
	testCase := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) service.ArticleService
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		wantBody   Result[int64]
	}{
		{
			name: "发布新建的文章",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(int64(1), nil)
				return svc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/articles/publish",
					bytes.NewReader([]byte(`
							{
							"id": 0,
							"content": "test content",
							"title": "test title"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: Result[int64]{Code: 0, Msg: "发布成功", Data: 1},
		},
		{
			name: "发布已有的文章",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(int64(3), nil)
				return svc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/articles/publish",
					bytes.NewReader([]byte(`
							{
							"id": 3,
							"content": "test content",
							"title": "test title"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: Result[int64]{Code: 0, Msg: "发布成功", Data: 3},
		},
		{
			name: "发布文章失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("发布失败"))
				return svc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/articles/publish",
					bytes.NewReader([]byte(`
							{
							"id": 3,
							"content": "test content",
							"title": "test title"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: Result[int64]{Code: 5, Msg: "系统错误", Data: 0},
		},
		{
			name: "参数Bind失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/articles/publish",
					bytes.NewReader([]byte(`
							{
							"id": "2342",
							"content": "test content",
							"title": "test title"}
							`)))
				if err != nil {
					panic("test case prepare failed:" + err.Error())
				}
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				return req
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			h := NewArticleHandler(svc, logger.NewNopLogger())
			server := gin.Default()
			server.Use(func(c *gin.Context) {
				c.Set("user", ijwt.TokenClaims{Uid: 123})
			})
			h.RegisterRoutes(server)
			req := tc.reqBuilder(t)

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Result[int64]
			err := json.NewDecoder(w.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}
