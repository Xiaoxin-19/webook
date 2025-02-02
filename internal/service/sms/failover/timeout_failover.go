package failover

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"syscall"
	"webok/internal/service/sms"
)

type TimeoutFailOverSMSService struct {
	services  []sms.Service
	idx       uint64
	cnt       uint32
	threshold uint32
}

func (t *TimeoutFailOverSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	cnt := atomic.LoadUint32(&t.cnt)
	curIdx := atomic.LoadUint64(&t.idx)

	if cnt >= t.threshold {
		newIdx := curIdx + 1
		if atomic.CompareAndSwapUint64(&t.idx, curIdx, newIdx) {
			atomic.StoreUint32(&t.cnt, 0)
		}
		curIdx = newIdx
	}

	svc := t.services[curIdx%uint64(len(t.services))]
	err := svc.Send(ctx, tplId, args, number...)

	switch {
	case err == nil:
		// 成功发送，重置计数器
		atomic.StoreUint32(&t.cnt, 0)
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		// 超时错误，增加计数器
		atomic.AddUint32(&t.cnt, 1)
	default:
		// 其他错误，考虑是否需要切换服务实例
		// 例如，如果是网络错误，可以直接切换
		if isNetworkError(err) {
			if atomic.CompareAndSwapUint64(&t.idx, curIdx, curIdx+1) {
				atomic.StoreUint32(&t.cnt, 0)
			}
		}
		return err
	}
	return err
}

func isNetworkError(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET)
}
func NewTimeoutFailOverSMSService(services []sms.Service, t uint32) *TimeoutFailOverSMSService {
	return &TimeoutFailOverSMSService{
		services:  services,
		threshold: t,
	}
}
