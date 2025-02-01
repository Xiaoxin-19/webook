package ratelimit

import (
	"context"
	"errors"
	"webok/internal/service/sms"
	"webok/pkg/limiter"
)

var ErrLimited = errors.New("触发限流")

type LimitSMSService struct {
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func (r *LimitSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		return ErrLimited
	}
	return r.svc.Send(ctx, tplId, args, number...)
}

func NewRateLimitSMSService(service sms.Service, l limiter.Limiter) *LimitSMSService {
	return &LimitSMSService{
		svc:     service,
		limiter: l,
		key:     "sms-limiter",
	}
}
