package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = s.appId
	request.SetContext(ctx)
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr[string](tplId)
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(number)

	response, err := s.client.SendSms(request)
	if err != nil {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}

	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			//不可能进入这里
			continue
		}
		if statusPtr.Code == nil || *statusPtr.Code != "Ok" {
			return fmt.Errorf("failed SendSms code: %s, msg:%s", *statusPtr.Code, *statusPtr.Message)
		}
	}
	return nil
}

func (s *Service) toPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data, func(idx int, src string) *string {
		return &src
	})
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &signName,
	}
}
