package localsms

import (
	"context"
	"log"
)

type Service struct {
}

func NewLocalSmsService() *Service {
	return &Service{}
}

func (s *Service) Send(_ context.Context, _ string, args []string, _ ...string) error {
	log.Println("验证码是", args)
	return nil
}
