package ioc

import (
	"os"
	"webok/internal/service/outh2/wechat"
)

func InitWechatService() wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到微信的 app id")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到微信的 app secret")
	}
	return wechat.NewService(appID, appSecret)
}
