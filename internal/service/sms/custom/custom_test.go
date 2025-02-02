package custom

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webok/internal/service/sms"
	smsmocks "webok/internal/service/sms/mock"
)

func TestCustomSMSService_Send(t *testing.T) {
	testCase := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) []sms.Service
		cnt   []int32
		reTry []int32
		idx   int32

		wantCnt   []int32
		wantReTry []int32
		wantIdx   uint64
		wantErr   error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{0, 0},
			reTry: []int32{0, 0},
			idx:   0,

			wantCnt:   []int32{0, 0},
			wantReTry: []int32{0, 0},
			wantIdx:   1,
			wantErr:   nil,
		},
		{
			name: "发送成功，但耗时超过最大耗时，更新错误追踪数据",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string, _ []string, _ ...string) error {
					time.Sleep(1*time.Second + 100*time.Millisecond)
					return nil
				})
				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{0, 0},
			reTry: []int32{0, 0},
			idx:   0,

			wantCnt:   []int32{0, 1},
			wantReTry: []int32{0, 0},
			wantIdx:   1,
			wantErr:   nil,
		},
		{
			name: "首次发送超时失败，更新错误追踪数据，重试成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.DeadlineExceeded)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{0, 0},
			reTry: []int32{0, 0},
			idx:   0,

			wantCnt:   []int32{0, 1},
			wantReTry: []int32{0, 0},
			wantIdx:   1,
			wantErr:   nil,
		},
		{
			name: "重试次数超过阈值，重启服务，发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{3, 3},
			reTry: []int32{10, 5},
			idx:   0,

			wantCnt:   []int32{0, 3},
			wantReTry: []int32{0, 6},
			wantIdx:   1,
			wantErr:   nil,
		},
		{
			name: "重试次数超过阈值，重启服务，服务发送失败，重试成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{3, 0},
			reTry: []int32{10, 0},
			idx:   1,

			wantCnt:   []int32{4, 0},
			wantReTry: []int32{0, 0},
			wantIdx:   2,
			wantErr:   nil,
		},
		{
			name: "无服务可用",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)

				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{3, 3},
			reTry: []int32{5, 5},
			idx:   0,

			wantCnt:   []int32{3, 3},
			wantReTry: []int32{6, 6},
			wantIdx:   1,
			wantErr:   ErrNoServiceAvailable,
		},
		{
			name: "上下文取消,不进行重试操作",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(context.Canceled)
				return []sms.Service{svc0, svc1}
			},
			cnt:   []int32{0, 0},
			reTry: []int32{0, 0},
			idx:   0,

			wantCnt:   []int32{0, 0},
			wantReTry: []int32{0, 0},
			wantIdx:   1,
			wantErr:   context.Canceled,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewCustomSMSService(tc.mock(ctrl), 3, 10, time.Second)
			svc.cnt = tc.cnt
			svc.reTryTimes = tc.reTry
			svc.idx = uint64(tc.idx)

			gotError := svc.Send(context.Background(), "123", []string{"123456"}, "15212345678")
			assert.Equal(t, tc.wantErr, gotError)
			assert.Equal(t, tc.wantCnt, svc.cnt)
			assert.Equal(t, tc.wantReTry, svc.reTryTimes)
			assert.Equal(t, tc.wantIdx, svc.idx)
		})
	}
}
