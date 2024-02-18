package ioc

import (
	"example/wb/internal/service/sms"
	"example/wb/internal/service/sms/localsms"
)

func InitSMSService() sms.Service {
	return localsms.NewLocalService()

}
