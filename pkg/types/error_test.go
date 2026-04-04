package types

import (
	"encoding/xml"
	"testing"
)

func TestRespErrorHasError(t *testing.T) {
	tests := []struct {
		name     string
		err      RespError
		expected bool
	}{
		{
			name: "res_code non-zero int",
			err: RespError{
				ResCode:    1,
				ResMessage: "error message",
			},
			expected: true,
		},
		{
			name: "res_code zero",
			err: RespError{
				ResCode:    0,
				ResMessage: "",
			},
			expected: false,
		},
		{
			name: "res_code string non-empty",
			err: RespError{
				ResCode:    "ERROR_CODE",
				ResMessage: "error",
			},
			expected: true,
		},
		{
			name: "Code field non-SUCCESS",
			err: RespError{
				Code:    "FAILURE",
				Message: "failed",
			},
			expected: true,
		},
		{
			name: "Code field SUCCESS",
			err: RespError{
				Code:    "SUCCESS",
				Message: "ok",
			},
			expected: false,
		},
		{
			name: "ErrorCode non-empty",
			err: RespError{
				ErrorCode: "ERR001",
				ErrorMsg:  "error occurred",
			},
			expected: true,
		},
		{
			name: "Error_ non-empty",
			err: RespError{
				Error_:  "invalid_request",
				Message: "invalid request",
			},
			expected: true,
		},
		{
			name:     "empty error",
			err:      RespError{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.HasError()
			if result != tt.expected {
				t.Errorf("HasError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRespErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      RespError
		expected string
	}{
		{
			name: "res_code error",
			err: RespError{
				ResCode:    1,
				ResMessage: "res error message",
			},
			expected: "res error message",
		},
		{
			name: "Code field error with Msg",
			err: RespError{
				Code: "FAILURE",
				Msg:  "failure message",
			},
			expected: "failure message",
		},
		{
			name: "Code field error with Message",
			err: RespError{
				Code:    "ERROR",
				Message: "error message",
			},
			expected: "error message",
		},
		{
			name: "Code field error without message",
			err: RespError{
				Code: "ERROR_CODE",
			},
			expected: "ERROR_CODE",
		},
		{
			name: "ErrorCode error",
			err: RespError{
				ErrorCode: "ERR001",
				ErrorMsg:  "error description",
			},
			expected: "error description",
		},
		{
			name: "Error_ field",
			err: RespError{
				Error_:  "invalid",
				Message: "invalid request",
			},
			expected: "invalid request",
		},
		{
			name:     "no error",
			err:      RespError{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRespErrorXMLParsing(t *testing.T) {
	err := RespError{
		XMLName: xml.Name{Local: "error"},
		Code:    "ERROR001",
		Message: "Test error message",
	}

	if err.Code != "ERROR001" {
		t.Errorf("XML Code = %s, want ERROR001", err.Code)
	}

	if err.Message != "Test error message" {
		t.Errorf("XML Message = %s, want 'Test error message'", err.Message)
	}
}

func TestCloudError(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
	}{
		{
			name:    "simple error",
			code:    "ERR001",
			message: "Simple error message",
		},
		{
			name:    "empty code",
			code:    "",
			message: "Error without code",
		},
		{
			name:    "empty message",
			code:    "ERR002",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCloudError(tt.code, tt.message)

			if err.Code != tt.code {
				t.Errorf("Code = %s, want %s", err.Code, tt.code)
			}

			if err.Message != tt.message {
				t.Errorf("Message = %s, want %s", err.Message, tt.message)
			}

			if err.Error() != tt.message {
				t.Errorf("Error() = %s, want %s", err.Error(), tt.message)
			}
		})
	}
}

func TestNewCloudErrorWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field":  "username",
		"reason": "too short",
	}

	err := NewCloudErrorWithDetails("VALIDATION_ERROR", "Validation failed", details)

	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %s, want VALIDATION_ERROR", err.Code)
	}

	if err.Message != "Validation failed" {
		t.Errorf("Message = %s, want 'Validation failed'", err.Message)
	}

	if err.Details["field"] != "username" {
		t.Errorf("Details[field] = %v, want 'username'", err.Details["field"])
	}
}
