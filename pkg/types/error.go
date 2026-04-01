package types

import "encoding/xml"

type RespError struct {
	ResCode    interface{} `json:"res_code"`
	ResMessage string      `json:"res_message"`

	Error_ string `json:"error"`

	XMLName   xml.Name `xml:"error"`
	Code      string   `json:"code" xml:"code"`
	Message   string   `json:"message" xml:"message"`
	Msg       string   `json:"msg"`
	ErrorCode string   `json:"errorCode"`
	ErrorMsg  string   `json:"errorMsg"`
}

func (e *RespError) HasError() bool {
	switch v := e.ResCode.(type) {
	case int, int64, int32:
		return v != 0
	case string:
		return e.ResCode != ""
	}
	return (e.Code != "" && e.Code != "SUCCESS") || e.ErrorCode != "" || e.Error_ != ""
}

func (e *RespError) Error() string {
	switch v := e.ResCode.(type) {
	case int, int64, int32:
		if v != 0 {
			return e.ResMessage
		}
	case string:
		if e.ResCode != "" {
			return e.ResMessage
		}
	}

	if e.Code != "" && e.Code != "SUCCESS" {
		if e.Msg != "" {
			return e.Msg
		}
		if e.Message != "" {
			return e.Message
		}
		return e.Code
	}

	if e.ErrorCode != "" {
		return e.ErrorMsg
	}

	if e.Error_ != "" {
		return e.Message
	}
	return ""
}

type CloudError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *CloudError) Error() string {
	return e.Message
}

func NewCloudError(code, message string) *CloudError {
	return &CloudError{
		Code:    code,
		Message: message,
	}
}

func NewCloudErrorWithDetails(code, message string, details map[string]interface{}) *CloudError {
	return &CloudError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
