package failover

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webok/internal/service/sms"
	smsmocks "webok/internal/service/sms/mock"
)

func TestTimeoutFailOverSMSService_Send(t *testing.T) {
	testCase := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) []sms.Service
		idx     uint64
		cnt     uint32
		wantIdx uint64
		wantCnt uint32
		wantErr error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0}
			},
			wantErr: nil,
		},
		{
			name: "触发切换-切换后成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			idx: 0,
			cnt: 3,
			// 触发了切换
			wantIdx: 1,
			wantCnt: 0,
			wantErr: nil,
		},
		{
			name: "触发切换-切换后依然失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				return []sms.Service{svc0, svc1}
			},
			idx:     1,
			cnt:     3,
			wantIdx: 2,
			wantCnt: 0,
			wantErr: errors.New("发送失败"),
		},
		{
			name: "触发切换-超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).
					Return(context.DeadlineExceeded)
				return []sms.Service{svc0, svc1}
			},
			idx:     1,
			cnt:     3,
			wantIdx: 2,
			wantCnt: 1,
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewTimeoutFailOverSMSService(tc.mock(ctrl), 3)
			svc.idx = tc.idx
			svc.cnt = tc.cnt
			gotErr := svc.Send(context.Background(), "123", []string{"123456"}, "15212345678")
			assert.Equal(t, tc.wantErr, gotErr)
			assert.Equal(t, tc.wantCnt, svc.cnt)
			assert.Equal(t, tc.wantIdx, svc.idx)
		})
	}
}
