package sms

import (
	"context"
)

// Service 发送短信的抽象
// 为了屏蔽不同的第三方实现
//
//go:generate mockgen -source=types.go -package=smsmocks -destination=./mock/sms.mock.go
type Service interface {
	Send(ctx context.Context, tplId string, args []string, number ...string) error
}
