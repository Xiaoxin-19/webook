package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"net/url"
	"webok/internal/domain"
)

type Service interface {
	AuthURL(ctx context.Context) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

var redirectURL = url.PathEscape("https://xiaoxina.xyz/oauth2/wechat/callback")

const authURLPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"

const accessTokenUrlPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"

type service struct {
	appID     string
	appSecret string
	client    *http.Client
}

type Result struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	UnionId      string `json:"unionid"`
	Scope        string `json:"scope"`

	// err response
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewService(appid string, appSecret string) Service {
	return &service{
		appID:     appid,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}
func (s *service) AuthURL(ctx context.Context) (string, error) {
	state := uuid.New()
	retStr := fmt.Sprintf(authURLPattern, s.appID, redirectURL, state)
	return retStr, nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenUrl := fmt.Sprintf(accessTokenUrlPattern, s.appID, s.appSecret, code)
	// 发送请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenUrl, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	httpResp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	// 解析结果
	var res Result
	err = json.NewDecoder(httpResp.Body).Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("调用微信接口失败 errcode %d, errmsg %s", res.ErrCode, res.ErrMsg)
	}
	return domain.WechatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil
}
