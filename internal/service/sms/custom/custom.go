package custom

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"
	"webok/internal/service/sms"
	"webok/internal/service/sms/failover"
)

var ErrNoServiceAvailable = failover.ErrNoServiceAvailable

type CustomSMSService struct {
	services       []sms.Service
	reTryThreshold int32
	threshold      int32
	maxCost        time.Duration
	idx            uint64
	cnt            []int32
	reTryTimes     []int32
}

func (c *CustomSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	var err error
	curIdx := atomic.AddUint64(&c.idx, 1)
	svcNum := uint64(len(c.services))
	for i := curIdx; i < curIdx+svcNum; i++ {
		idx := i % svcNum
		curCnt := atomic.LoadInt32(&c.cnt[idx])

		// 判断当前服务是否可用，是否需要重置服务可用状态
		if curCnt >= c.threshold {
			reTry := atomic.LoadInt32(&c.reTryTimes[idx])
			if reTry < c.reTryThreshold {
				atomic.AddInt32(&c.reTryTimes[idx], 1)
				continue
			}
			// 超过重试阈值恢复可以用状态，
			atomic.StoreInt32(&c.reTryTimes[idx], 0)
		}

		svc := c.services[idx]
		startTime := time.Now()
		err = svc.Send(ctx, tplId, args, number...)
		costTime := time.Since(startTime)
		switch {
		case costTime > c.maxCost:
			atomic.AddInt32(&c.cnt[idx], 1)
			return nil
		case err == nil:
			// 成功发送，重置计数器
			atomic.StoreInt32(&c.cnt[idx], 0)
			return nil
		case errors.Is(err, context.DeadlineExceeded):
			atomic.AddInt32(&c.cnt[idx], 1)
		case errors.Is(err, context.Canceled):
			return err
		default:
			log.Printf("unknow err : %v\n", err)
		}
	}
	return ErrNoServiceAvailable
}

func NewCustomSMSService(s []sms.Service, t uint32, reTry uint32, maxCost time.Duration) *CustomSMSService {
	return &CustomSMSService{
		services:       s,
		threshold:      int32(t),
		reTryThreshold: int32(reTry),
		maxCost:        maxCost,
		cnt:            make([]int32, len(s)),
	}
}
