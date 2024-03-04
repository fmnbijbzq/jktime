package ioc

import (
	"example/wb/internal/service/oauth2/wechat"
	"log"
	"os"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		appId = "213"
		log.Println("wechat 环境变量未找到")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		appSecret = "231"
		log.Println("wechat 环境变量未找到")
	}
	wc := wechat.NewService(appId, appSecret)
	return wc
}
