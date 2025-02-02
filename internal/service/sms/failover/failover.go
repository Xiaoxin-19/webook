package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webok/internal/service/sms"
)

var (
	ErrNoServiceAvailable = errors.New("无服务可用")
)

type FailOverSmsService struct {
	services []sms.Service
}

type FailOverV2SmsService struct {
	services []sms.Service
	idx      uint64
}

func (f *FailOverSmsService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	for _, svc := range f.services {
		err := svc.Send(ctx, tplId, args, number...)
		if err != nil {
			// 打印相关日志
			log.Printf("failover failover service %s error: %v", tplId, err)
			continue
		}
		return nil
	}
	return ErrNoServiceAvailable
}

func (f *FailOverV2SmsService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	starIdx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.services))
	for i := starIdx; i < starIdx+length; i++ {
		err := f.services[i%length].Send(ctx, tplId, args, number...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return err
		}
	}
	return ErrNoServiceAvailable
}

func NewFailOverSMSService(service []sms.Service) *FailOverSmsService {
	return &FailOverSmsService{
		services: service,
	}
}
func NewFailOverV2SMSService(service []sms.Service) *FailOverV2SmsService {
	return &FailOverV2SmsService{
		services: service,
	}
}
