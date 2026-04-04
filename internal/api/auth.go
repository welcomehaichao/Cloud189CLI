package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/skip2/go-qrcode"
	"github.com/yuhaichao/cloud189-cli/internal/crypto"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
	"github.com/yuhaichao/cloud189-cli/pkg/utils"
)

type LoginParam struct {
	Lt           string
	ReqId        string
	ParamId      string
	CaptchaToken string
	RsaUsername  string
	RsaPassword  string
	jRsaKey      string
}

type EncryptConfResp struct {
	Result int `json:"result"`
	Data   struct {
		UpSmsOn   string `json:"upSmsOn"`
		Pre       string `json:"pre"`
		PreDomain string `json:"preDomain"`
		PubKey    string `json:"pubKey"`
	} `json:"data"`
}

type LoginResp struct {
	Msg    string `json:"msg"`
	Result int    `json:"result"`
	ToUrl  string `json:"toUrl"`
}

func (c *Client) Login(username, password string) (*types.Session, error) {
	loginParam, err := c.initLoginParam(username, password)
	if err != nil {
		return nil, err
	}

	loginResp, err := c.submitLogin(loginParam)
	if err != nil {
		return nil, err
	}

	if loginResp.Result != 0 {
		return nil, fmt.Errorf("login failed: %s", loginResp.Msg)
	}

	session, err := c.getSession(loginResp.ToUrl)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (c *Client) initLoginParam(username, password string) (*LoginParam, error) {
	res, err := c.client.R().SetQueryParams(map[string]string{
		"appId":      AppID,
		"clientType": ClientType,
		"returnURL":  ReturnURL,
		"timeStamp":  fmt.Sprint(utils.Timestamp()),
	}).Get(WebURL + "/api/portal/unifyLoginForPC.action")

	if err != nil {
		return nil, err
	}

	html := res.String()

	captchaToken := regexp.MustCompile(`'captchaToken' value='(.+?)'`).FindStringSubmatch(html)
	lt := regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(html)
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(html)
	reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(html)

	if len(captchaToken) < 2 || len(lt) < 2 || len(paramId) < 2 || len(reqId) < 2 {
		return nil, fmt.Errorf("failed to parse login parameters")
	}

	param := &LoginParam{
		Lt:           lt[1],
		ReqId:        reqId[1],
		ParamId:      paramId[1],
		CaptchaToken: captchaToken[1],
	}

	var encryptConf EncryptConfResp
	res, err = c.client.R().
		ForceContentType("application/json;charset=UTF-8").
		SetFormData(map[string]string{"appId": AppID}).
		SetResult(&encryptConf).
		Post(AuthURL + "/api/logbox/config/encryptConf.do")

	if err != nil {
		return nil, err
	}

	if encryptConf.Result != 0 {
		return nil, fmt.Errorf("failed to get encrypt config")
	}

	publicKey := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", encryptConf.Data.PubKey)

	param.jRsaKey = publicKey
	param.RsaUsername = encryptConf.Data.Pre + mustRsaEncrypt(publicKey, username)
	param.RsaPassword = encryptConf.Data.Pre + mustRsaEncrypt(publicKey, password)

	return param, nil
}

func (c *Client) submitLogin(param *LoginParam) (*LoginResp, error) {
	var loginResp LoginResp
	_, err := c.client.R().
		ForceContentType("application/json;charset=UTF-8").
		SetHeaders(map[string]string{
			"REQID": param.ReqId,
			"lt":    param.Lt,
		}).
		SetFormData(map[string]string{
			"appKey":       AppID,
			"accountType":  AccountType,
			"userName":     param.RsaUsername,
			"password":     param.RsaPassword,
			"validateCode": "",
			"captchaToken": param.CaptchaToken,
			"returnUrl":    ReturnURL,
			"dynamicCheck": "FALSE",
			"clientType":   ClientType,
			"cb_SaveName":  "1",
			"isOauth2":     "false",
			"state":        "",
			"paramId":      param.ParamId,
		}).
		SetResult(&loginResp).
		Post(AuthURL + "/api/logbox/oauth2/loginSubmit.do")

	if err != nil {
		return nil, err
	}

	return &loginResp, nil
}

func (c *Client) getSession(toUrl string) (*types.Session, error) {
	var session types.Session
	var erron types.RespError

	_, err := c.client.R().
		SetQueryParams(c.ClientSuffix()).
		SetQueryParam("redirectURL", toUrl).
		SetHeader("X-Request-ID", utils.GenerateUUID()).
		SetResult(&session).
		SetError(&erron).
		Post(APIURL + "/getSessionForPC.action")

	if err != nil {
		return nil, err
	}

	if erron.HasError() {
		return nil, &erron
	}

	if session.ResCode != 0 {
		return nil, fmt.Errorf(session.ResMessage)
	}

	return &session, nil
}

func (c *Client) RefreshToken(refreshToken string) (*types.Session, error) {
	var session types.Session
	var erron types.RespError

	_, err := c.client.R().
		SetFormData(map[string]string{
			"clientId":     AppID,
			"refreshToken": refreshToken,
			"grantType":    "refresh_token",
			"format":       "json",
		}).
		SetResult(&session).
		SetError(&erron).
		Post(AuthURL + "/api/oauth2/refreshToken.do")

	if err != nil {
		return nil, err
	}

	if erron.HasError() {
		return nil, &erron
	}

	return &session, nil
}

func (c *Client) KeepAlive() error {
	_, err := c.Get(APIURL+"/keepUserSession.action", func(r *resty.Request) {
		r.SetQueryParams(c.ClientSuffix())
	}, false)

	return err
}

func (c *Client) RefreshSession(accessToken string) (*types.Session, error) {
	var session types.Session
	var erron types.RespError

	_, err := c.client.R().
		SetQueryParams(c.ClientSuffix()).
		SetQueryParams(map[string]string{
			"appId":       AppID,
			"accessToken": accessToken,
		}).
		SetHeader("X-Request-ID", utils.GenerateUUID()).
		SetResult(&session).
		SetError(&erron).
		Get(APIURL + "/getSessionForPC.action")

	if err != nil {
		return nil, err
	}

	if erron.HasError() {
		return nil, &erron
	}

	return &session, nil
}

func mustRsaEncrypt(publicKey, data string) string {
	encrypted, err := crypto.RsaEncrypt(publicKey, data)
	if err != nil {
		panic(err)
	}
	return encrypted
}

func (c *Client) initBaseParams() (*types.BaseLoginParam, error) {
	res, err := c.client.R().
		SetQueryParams(map[string]string{
			"appId":      AppID,
			"clientType": ClientType,
			"returnURL":  ReturnURL,
			"timeStamp":  fmt.Sprint(utils.Timestamp()),
		}).
		Get(WebURL + "/api/portal/unifyLoginForPC.action")
	if err != nil {
		return nil, err
	}

	html := res.String()

	captchaToken := regexp.MustCompile(`'captchaToken' value='(.+?)'`).FindStringSubmatch(html)
	lt := regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(html)
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(html)
	reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(html)

	if len(captchaToken) < 2 || len(lt) < 2 || len(paramId) < 2 || len(reqId) < 2 {
		return nil, fmt.Errorf("failed to parse login parameters")
	}

	return &types.BaseLoginParam{
		CaptchaToken: captchaToken[1],
		Lt:           lt[1],
		ParamId:      paramId[1],
		ReqId:        reqId[1],
	}, nil
}

func (c *Client) initQRCodeParam() (*types.QRLoginParam, error) {
	baseParam, err := c.initBaseParams()
	if err != nil {
		return nil, err
	}

	var qrcodeParam types.QRLoginParam
	var erron types.RespError

	_, err = c.client.R().
		SetFormData(map[string]string{"appId": AppID}).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&qrcodeParam).
		SetError(&erron).
		Post(AuthURL + "/api/logbox/oauth2/getUUID.do")

	if err != nil {
		return nil, err
	}

	if erron.HasError() {
		return nil, &erron
	}

	qrcodeParam.BaseLoginParam = *baseParam
	return &qrcodeParam, nil
}

func (c *Client) checkQRCodeStatus(param *types.QRLoginParam) (int, string, error) {
	var state struct {
		Status      int    `json:"status"`
		RedirectUrl string `json:"redirectUrl"`
		Msg         string `json:"msg"`
	}
	var erron types.RespError

	now := time.Now()

	_, err := c.client.R().
		SetHeaders(map[string]string{
			"Referer": AuthURL,
			"Reqid":   param.ReqId,
			"lt":      param.Lt,
		}).
		SetFormData(map[string]string{
			"appId":      AppID,
			"clientType": ClientType,
			"returnUrl":  ReturnURL,
			"paramId":    param.ParamId,
			"uuid":       param.UUID,
			"encryuuid":  param.EncryUUID,
			"date":       utils.FormatDate(now),
			"timeStamp":  fmt.Sprint(now.UTC().UnixNano() / 1e6),
		}).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&state).
		SetError(&erron).
		Post(AuthURL + "/api/logbox/oauth2/qrcodeLoginState.do")

	if err != nil {
		return 0, "", fmt.Errorf("failed to check QR code status: %w", err)
	}

	if erron.HasError() {
		return 0, "", fmt.Errorf("API error: %s", erron.Error())
	}

	return state.Status, state.RedirectUrl, nil
}

func (c *Client) LoginByQRCode() (*types.Session, error) {
	param, err := c.initQRCodeParam()
	if err != nil {
		return nil, fmt.Errorf("failed to init QR code: %w", err)
	}

	fmt.Println("\n========== 天翼云盘二维码登录 ==========")
	fmt.Println()

	if err := printQRCode(param.UUID); err != nil {
		fmt.Printf("二维码链接: %s\n", param.UUID)
	} else {
		fmt.Println()
		fmt.Printf("或直接访问: %s\n", param.UUID)
	}

	fmt.Println()
	fmt.Println("请使用天翼云盘APP扫描上方二维码")
	fmt.Println("======================================")
	fmt.Println()

	for i := 0; i < 60; i++ {
		status, redirectUrl, err := c.checkQRCodeStatus(param)
		if err != nil {
			return nil, fmt.Errorf("failed to check status: %w", err)
		}

		switch status {
		case 0:
			if redirectUrl == "" {
				return nil, fmt.Errorf("no redirect URL received")
			}
			fmt.Println("\n✓ 登录成功！")
			return c.getSession(redirectUrl)
		case -106:
			if i == 0 {
				fmt.Print("\r⏳ 等待扫码...")
			}
		case -11002:
			fmt.Printf("\r📱 已扫码，等待确认... (%d秒)   ", (i+1)*5)
		case -11001:
			return nil, fmt.Errorf("二维码已过期，请重新登录")
		default:
			return nil, fmt.Errorf("登录失败，状态码: %d", status)
		}

		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("登录超时，请重试")
}

func printQRCode(content string) error {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return err
	}

	bitmap := qr.Bitmap()
	size := len(bitmap)

	border := 1

	for i := 0; i < border; i++ {
		fmt.Println(strings.Repeat(" ", size+2*border))
	}

	for y := 0; y < size; y += 2 {
		line := strings.Repeat(" ", border)
		for x := 0; x < size; x++ {
			upper := y < size && bitmap[y][x]
			lower := y+1 < size && bitmap[y+1][x]

			if upper && lower {
				line += "█"
			} else if upper {
				line += "▀"
			} else if lower {
				line += "▄"
			} else {
				line += " "
			}
		}
		line += strings.Repeat(" ", border)
		fmt.Println(line)
	}

	for i := 0; i < border; i++ {
		fmt.Println(strings.Repeat(" ", size+2*border))
	}

	return nil
}
