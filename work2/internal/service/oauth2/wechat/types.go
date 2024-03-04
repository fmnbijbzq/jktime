package wechat

import (
	"context"
	"encoding/json"
	"example/wb/internal/domain"
	"fmt"
	"net/http"
	"net/url"
)

type Service interface {
	AUthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

var redirectURL = url.PathEscape(`https://`)

const authURLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`

type service struct {
	appId     string
	appSecret string
	client    *http.Client
}

func NewService(appId string, appSecret string) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (s *service) AUthURL(ctx context.Context, state string) (string, error) {
	return fmt.Sprintf(authURLPattern, s.appId, redirectURL, state), nil

}
func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenUrl := fmt.Sprintf(`https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`,
		s.appId, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenUrl, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	httpResp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	var res Result
	err = json.NewDecoder(httpResp.Body).Decode(&res)
	if err != nil {
		// 转Json出错
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{},
			fmt.Errorf("调用微信接口失败 errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
	}

	return domain.WechatInfo{
		Unionid: res.Unionid,
		Openid:  res.Openid,
	}, nil

}

type Result struct {
	// 接口调用凭证
	AccessToken string
	// access_token接口调用凭证超时时间，单位（秒）
	ExpiresIn int64 `json:"expires_in"`
	// 用户刷新access_token
	RefreshToken string `json:"expires_in"`
	// 授权用户唯一标识
	Openid string `json:"openid"`
	// 用户授权的作用域，使用逗号（,）分隔
	Scope string `json:""scope`
	// 当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段。
	Unionid string `json:"unionid"`
	// 错误返回
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
