package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/yuhaichao/cloud189-cli/internal/config"
	"github.com/yuhaichao/cloud189-cli/internal/crypto"
	"github.com/yuhaichao/cloud189-cli/pkg/types"
	"github.com/yuhaichao/cloud189-cli/pkg/utils"
)

const (
	AccountType = "02"
	AppID       = "8025431004"
	ClientType  = "10020"
	Version     = "6.2"
	PC          = "TELEPC"
	ChannelID   = "web_cloud.189.cn"

	WebURL    = "https://cloud.189.cn"
	AuthURL   = "https://open.e.189.cn"
	APIURL    = "https://api.cloud.189.cn"
	UploadURL = "https://upload.cloud.189.cn"

	ReturnURL = "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html"
)

type Params map[string]string

func (p Params) Set(key, value string) {
	p[key] = value
}

func (p Params) Encode() string {
	if p == nil {
		return ""
	}
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(p[k])
	}
	return buf.String()
}

func (c *Client) EncryptParams(params Params, isFamily bool) string {
	sessionSecret := c.config.SessionSecret
	if isFamily {
		sessionSecret = c.config.FamilySessionSecret
	}
	if params != nil && len(sessionSecret) >= 16 {
		encrypted, err := crypto.AesEncryptHex([]byte(params.Encode()), []byte(sessionSecret[:16]))
		if err != nil {
			return ""
		}
		return encrypted
	}
	return ""
}

type Client struct {
	client  *resty.Client
	config  *config.Config
	manager *config.Manager
}

func NewClient(cfg *config.Config) *Client {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetHeaders(map[string]string{
			"Accept":  "application/json;charset=UTF-8",
			"Referer": WebURL,
		})

	return &Client{
		client: client,
		config: cfg,
	}
}

func NewClientWithManager(manager *config.Manager) *Client {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetHeaders(map[string]string{
			"Accept":  "application/json;charset=UTF-8",
			"Referer": WebURL,
		})

	return &Client{
		client:  client,
		config:  manager.GetConfig(),
		manager: manager,
	}
}

func (c *Client) ClientSuffix() map[string]string {
	return map[string]string{
		"clientType": PC,
		"version":    Version,
		"channelId":  ChannelID,
		"rand":       fmt.Sprintf("%d_%d", utils.RandomRange(1, 99999), utils.RandomRange(1, 9999999999)),
	}
}

func (c *Client) AuthClientSuffix() map[string]string {
	return map[string]string{
		"clientType": ClientType,
		"version":    Version,
		"channelId":  ChannelID,
		"rand":       fmt.Sprintf("%d_%d", utils.RandomRange(1, 99999), utils.RandomRange(1, 9999999999)),
	}
}

func (c *Client) SignatureHeader(url, method string, isFamily bool) map[string]string {
	// 自动检查并刷新Session
	if err := c.checkAndRefreshSession(); err != nil {
		// 如果刷新失败，继续使用当前Session（可能会在请求时失败）
		// 不阻止请求执行
	}

	date := utils.HTTPTime()
	sessionKey := c.config.SessionKey
	sessionSecret := c.config.SessionSecret

	if isFamily {
		sessionKey = c.config.FamilySessionKey
		sessionSecret = c.config.FamilySessionSecret
	}

	signature := crypto.SignatureOfHmac(sessionSecret, sessionKey, method, url, date)

	return map[string]string{
		"Date":         date,
		"SessionKey":   sessionKey,
		"X-Request-ID": uuid.NewString(),
		"Signature":    signature,
	}
}

func (c *Client) checkAndRefreshSession() error {
	// 如果没有manager，无法刷新
	if c.manager == nil {
		return nil
	}

	// 检查是否需要刷新
	if !c.manager.NeedRefresh() {
		return nil
	}

	// 检查是否有RefreshToken
	if c.config.RefreshToken == "" && c.config.AccessToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// 尝试刷新Session（静默刷新，不输出日志）
	return c.refreshSession()
}

func (c *Client) refreshSession() error {
	// 方法1: 使用RefreshToken刷新
	if c.config.RefreshToken != "" {
		session, err := c.RefreshToken(c.config.RefreshToken)
		if err == nil && session != nil && session.SessionKey != "" {
			return c.updateSession(session)
		}
	}

	// 方法2: 使用AccessToken刷新
	if c.config.AccessToken != "" {
		session, err := c.RefreshSession(c.config.AccessToken)
		if err == nil && session != nil && session.SessionKey != "" {
			return c.updateSession(session)
		}
	}

	return fmt.Errorf("failed to refresh session")
}

func (c *Client) updateSession(session *types.Session) error {
	if c.manager == nil {
		return fmt.Errorf("no config manager available")
	}

	// 更新配置
	c.config.SessionKey = session.SessionKey
	c.config.SessionSecret = session.SessionSecret
	c.config.FamilySessionKey = session.FamilySessionKey
	c.config.FamilySessionSecret = session.FamilySessionSecret

	if session.AccessToken != "" {
		c.config.AccessToken = session.AccessToken
	}
	if session.RefreshToken != "" {
		c.config.RefreshToken = session.RefreshToken
	}

	// 更新过期时间
	c.config.ExpiresAt = time.Now().Add(24 * time.Hour)

	// 保存配置
	return c.manager.Save()
}

func (c *Client) SignatureHeaderWithParams(url, method, params string, isFamily bool) map[string]string {
	// 自动检查并刷新Session
	if err := c.checkAndRefreshSession(); err != nil {
		// 如果刷新失败，继续使用当前Session（可能会在请求时失败）
		// 不阻止请求执行
	}

	date := utils.HTTPTime()
	sessionKey := c.config.SessionKey
	sessionSecret := c.config.SessionSecret

	if isFamily {
		sessionKey = c.config.FamilySessionKey
		sessionSecret = c.config.FamilySessionSecret
	}

	signature := crypto.SignatureOfHmacWithParams(sessionSecret, sessionKey, method, url, date, params)

	return map[string]string{
		"Date":         date,
		"SessionKey":   sessionKey,
		"X-Request-ID": uuid.NewString(),
		"Signature":    signature,
	}
}

func (c *Client) RequestWithParams(url, method string, callback func(*resty.Request), params Params, isFamily bool) ([]byte, error) {
	req := c.client.R().SetQueryParams(c.ClientSuffix())

	paramsData := c.EncryptParams(params, isFamily)
	if paramsData != "" {
		req.SetQueryParam("params", paramsData)
	}

	req.SetHeaders(c.SignatureHeaderWithParams(url, method, paramsData, isFamily))

	if callback != nil {
		callback(req)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

func (c *Client) Request(url, method string, callback func(*resty.Request), isFamily bool) ([]byte, error) {
	req := c.client.R().SetQueryParams(c.ClientSuffix())

	req.SetHeaders(c.SignatureHeader(url, method, isFamily))

	if callback != nil {
		callback(req)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

func (c *Client) Get(url string, callback func(*resty.Request), isFamily bool) ([]byte, error) {
	return c.Request(url, http.MethodGet, callback, isFamily)
}

func (c *Client) Post(url string, callback func(*resty.Request), isFamily bool) ([]byte, error) {
	return c.Request(url, http.MethodPost, callback, isFamily)
}

func (c *Client) getClient() *resty.Client {
	return c.client
}
