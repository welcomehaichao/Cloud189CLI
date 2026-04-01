package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

func HmacSha1(data, secret string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func HmacSha1Upper(data, secret string) string {
	return strings.ToUpper(HmacSha1(data, secret))
}

func SignatureOfHmac(sessionSecret, sessionKey, operate, url, date string) string {
	urlpath := extractURLPath(url)
	data := fmt.Sprintf("SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s",
		sessionKey, operate, urlpath, date)
	return HmacSha1Upper(data, sessionSecret)
}

func SignatureOfHmacWithParams(sessionSecret, sessionKey, operate, url, date, params string) string {
	urlpath := extractURLPath(url)
	data := fmt.Sprintf("SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s&params=%s",
		sessionKey, operate, urlpath, date, params)
	return HmacSha1Upper(data, sessionSecret)
}

func AppKeySignatureOfHmac(appSignatureSecret, appKey, operate, url string, timestamp int64) string {
	urlpath := extractURLPath(url)
	data := fmt.Sprintf("AppKey=%s&Operate=%s&RequestURI=%s&Timestamp=%d",
		appKey, operate, urlpath, timestamp)
	return HmacSha1Upper(data, appSignatureSecret)
}

func extractURLPath(url string) string {
	re := regexp.MustCompile(`://[^/]+((/[^/\s?#]+)*)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
