package ratelimit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webok/internal/service/sms"
	smsmocks "webok/internal/service/sms/mock"
	"webok/pkg/limiter"
	limitmocks "webok/pkg/limiter/mock"
)

func TestLimitSMSServiceSend(t *testing.T) {
	testCase := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter)
		wantErr error
	}{
		{
			name: "没有触发限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				return svc, limit
			},
			wantErr: nil,
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, limit
			},
			wantErr: ErrLimited,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("redis error"))
				return svc, limit
			},
			wantErr: errors.New("redis error"),
		},
		{
			name: "限流器错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limit := limitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("limit error"))
				return svc, limit
			},
			wantErr: errors.New("limit error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			smsMock, limitMock := tt.mock(ctrl)

			limitSms := NewRateLimitSMSService(smsMock, limitMock)
			gotErr := limitSms.Send(context.Background(), "12312", []string{"1234"}, "15212345678")
			assert.Equal(t, tt.wantErr, gotErr)
		})
	}
}
