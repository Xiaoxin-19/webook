package sms

import (
	"context"
)

// Service 发送短信的抽象
// 为了屏蔽不同的第三方实现
type Service interface {
	Send(ctx context.Context, tplId string, args []string, number ...string) error
}
