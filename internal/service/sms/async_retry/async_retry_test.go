package async_retry

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webok/internal/service/sms"
	smsmocks "webok/internal/service/sms/mock"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestAsyncRetrySMSService_Send(t *testing.T) {
	reTryInterval := int64(3)
	testCase := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable)

		wantErr error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				//all expectations were already fulfilled, call to cmd '[exists async_retry_sms_task:1234567890]' was not expected
				rMock.Regexp().ExpectExists(`async_retry_sms_task:\d+`).SetVal(0)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return svc, rdb
			},
			wantErr: nil,
		},
		{
			name: "发送失败，任务已存在",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectExists(`async_retry_sms_task:\d+`).SetVal(1)
				return svc, rdb
			},
			wantErr: ErrSendFailedPendingRetry,
		},
		{
			name: "发送失败，存储任务元数据失败",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectExists(`async_retry_sms_task:\d+`).SetVal(0)

				rMock.ExpectTxPipeline()
				rMock.Regexp().ExpectHSet(".*", "tplId", ".*", "args", ".*", "numbers", ".*", "retries", ".*", "maxRetries", ".*").
					SetVal(1)
				rMock.Regexp().ExpectZAdd(AsyncQueueKey, redis.Z{
					Score:  float64(time.Now().Unix() + reTryInterval),
					Member: ".*",
				}).SetErr(errors.New("redis error"))
				rMock.ExpectTxPipelineExec()
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("send failed"))
				return svc, rdb
			},
			wantErr: errors.New("redis error"),
		},
		{
			name: "发送失败，存储任务元数据成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectExists(`async_retry_sms_task:\d+`).SetVal(0)

				rMock.ExpectTxPipeline()
				rMock.Regexp().ExpectHSet(".*", "tplId", ".*", "args", ".*", "numbers", ".*", "retries", ".*", "maxRetries", ".*").
					SetVal(1)
				rMock.Regexp().ExpectZAdd(AsyncQueueKey, redis.Z{
					Score:  float64(time.Now().Unix() + reTryInterval),
					Member: ".*",
				}).SetVal(1)
				rMock.ExpectTxPipelineExec()
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("send failed"))
				return svc, rdb
			},
			wantErr: ErrSendFailedPendingRetry,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc, rdb := tt.mock(ctrl)
			retryService := NewAsyncRetrySMSService(svc, rdb, 3, int64(reTryInterval), 6)
			err := retryService.Send(context.Background(), "tplId", []string{"arg1"}, "1234567890")
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAsyncRetrySMSService_AsyncRetryWorker(t *testing.T) {
	reTryInterval := int64(3)
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable)
		wantErr error
	}{
		{
			name: "异步重试-没有任务",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectZRangeByScore(AsyncQueueKey, &redis.ZRangeBy{
					Min: ".*",
					Max: ".*",
				}).SetVal([]string{})
				return svc, rdb
			},
			wantErr: nil,
		},
		{
			name: "异步重试-有任务-重试失败",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectZRangeByScore(AsyncQueueKey, &redis.ZRangeBy{
					Min: ".*",
					Max: ".*",
				}).SetVal([]string{"123"})
				rMock.ExpectTxPipeline()
				rMock.Regexp().ExpectHGetAll(".*").SetVal(map[string]string{
					"tplId":      "123",
					"args":       "[]",
					"numbers":    "[]",
					"retries":    "0",
					"maxRetries": "3",
				})
				rMock.Regexp().ExpectZRem(AsyncQueueKey, "123").SetVal(int64(1))
				rMock.ExpectTxPipelineExec()
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("send failed"))
				rMock.Regexp().ExpectHSet(".*", "retries", ".*").
					SetVal(1)
				rMock.Regexp().ExpectZAdd(AsyncQueueKey, redis.Z{
					Score:  float64(time.Now().Unix() + reTryInterval),
					Member: "123",
				}).SetVal(int64(1))

				return svc, rdb
			},
			wantErr: nil,
		},
		{
			name: "异步重试-有任务-超过最大重试次数",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectZRangeByScore(AsyncQueueKey, &redis.ZRangeBy{
					Min: ".*",
					Max: ".*",
				}).SetVal([]string{"123"})
				rMock.ExpectTxPipeline()
				rMock.Regexp().ExpectHGetAll(".*").SetVal(map[string]string{
					"tplId":      "123",
					"args":       "[]",
					"numbers":    "[]",
					"retries":    "3",
					"maxRetries": "3",
				})
				rMock.Regexp().ExpectZRem(AsyncQueueKey, "123").SetVal(int64(1))
				rMock.ExpectTxPipelineExec()
				return svc, rdb
			},
			wantErr: nil,
		},
		{
			name: "异步重试-有任务-重试成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, redis.Cmdable) {
				svc := smsmocks.NewMockService(ctrl)
				rdb, rMock := redismock.NewClientMock()
				rMock.Regexp().ExpectZRangeByScore(AsyncQueueKey, &redis.ZRangeBy{
					Min: ".*",
					Max: ".*",
				}).SetVal([]string{"123"})
				rMock.ExpectTxPipeline()
				rMock.Regexp().ExpectHGetAll(".*").SetVal(map[string]string{
					"tplId":      "123",
					"args":       `["arg1"]`,
					"numbers":    `["15212345678"]`,
					"retries":    "2",
					"maxRetries": "3",
				})
				rMock.Regexp().ExpectZRem(AsyncQueueKey, "123").SetVal(int64(1))
				rMock.ExpectTxPipelineExec()
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				//rMock.Regexp().ExpectDel(".*").SetVal(int64(1))
				rMock.Regexp().ExpectDel(".*").SetErr(errors.New("redis error"))
				return svc, rdb
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, rdb := tc.mock(ctrl)
			service := NewAsyncRetrySMSService(svc, rdb, 3, int64(reTryInterval), 6)
			gotErr := service.AsyncRetryWorker()
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}
