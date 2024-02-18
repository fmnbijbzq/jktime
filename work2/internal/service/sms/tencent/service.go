package tencent

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type SMSSerivce struct {
	client   *sms.Client
	appId    string
	signName string
}

func NewService(client *sms.Client, appId string, signName string) *SMSSerivce {
	return &SMSSerivce{
		client:   client,
		appId:    appId,
		signName: signName,
	}
}

func NewSMSClient() *sms.Client {
	crd := common.NewCredential("SecreId", "SecretKey")

	cpf := profile.NewClientProfile()

	cpf.HttpProfile.ReqMethod = "POST"

	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	client, _ := sms.NewClient(crd, "ap-guangzhou", cpf)
	return client

}

func (svc *SMSSerivce) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// crd := common.NewCredential("SecreId", "SecretKey")

	// cpf := profile.NewClientProfile()

	// cpf.HttpProfile.ReqMethod = "POST"

	// cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	// client, _ := sms.NewClient(crd, "ap-guangzhou", cpf)

	request := sms.NewSendSmsRequest()
	request.SetContext(ctx)
	// 应用ID
	request.SmsSdkAppId = common.StringPtr(svc.appId)
	/* 短信签名内容: 使用 UTF-8 编码，必须填写已审核通过的签名 */
	// 签名信息可前往 [国内短信](https://console.cloud.tencent.com/smsv2/csms-sign) 或 [国际/港澳台短信](https://console.cloud.tencent.com/smsv2/isms-sign) 的签名管理查看
	request.SignName = common.StringPtr(svc.signName)

	/* 模板 ID: 必须填写已审核通过的模板 ID */
	// 模板 ID 可前往 [国内短信](https://console.cloud.tencent.com/smsv2/csms-template) 或 [国际/港澳台短信](https://console.cloud.tencent.com/smsv2/isms-template) 的正文模板管理查看
	request.TemplateId = common.StringPtr(tplId)

	/* 模板参数: 模板参数的个数需要与 TemplateId 对应模板的变量个数保持一致，若无模板参数，则设置为空*/
	request.TemplateParamSet = common.StringPtrs(args)

	/* 下发手机号码，采用 E.164 标准，+[国家或地区码][手机号]
	 * 示例如：+8613711112222， 其中前面有一个+号 ，86为国家码，13711112222为手机号，最多不要超过200个手机号*/
	request.PhoneNumberSet = common.StringPtrs(numbers)

	/* 用户的 session 内容（无需要可忽略）: 可以携带用户侧 ID 等上下文信息，server 会原样返回 */
	// request.SessionContext = common.StringPtr("")

	/* 短信码号扩展号（无需要可忽略）: 默认未开通，如需开通请联系 [腾讯云短信小助手] */
	// request.ExtendCode = common.StringPtr("")

	/* 国内短信无需填写该项；国际/港澳台短信已申请独立 SenderId 需要填写该字段，默认使用公共 SenderId，无需填写该字段。注：月度使用量达到指定量级可申请独立 SenderId 使用，详情请联系 [腾讯云短信小助手](https://cloud.tencent.com/document/product/382/3773#.E6.8A.80.E6.9C.AF.E4.BA.A4.E6.B5.81)。 */
	// request.SenderId = common.StringPtr("")

	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := svc.client.SendSms(request)
	// 处理异常
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return nil
	}
	// 非SDK异常，直接失败。实际代码中可以加入其他的处理。
	if err != nil {
		fmt.Printf("An API error has returned: %s", err)
		return nil
	}
	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			// 不可能进来这里
			continue
		}
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败，code：%s，msg：%s", *status.Code, *status.Message)

		}

	}
	return nil
}
