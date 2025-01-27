package ioc

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"webok/internal/service/sms"
	"webok/internal/service/sms/localsms"
	"webok/internal/service/sms/tencent"
)

func InitSMSService() sms.Service {
	return localsms.NewLocalSmsService()
	//return InitTencentSmsService()
}

func InitTencentSmsService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("can not find secret Id for tencent")
	}

	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("can not find secret Key for tencent")
	}
	c, err := tencentSms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile(),
	)
	if err != nil {
		panic(err)
	}

	return tencent.NewService(c, "", "")
}
