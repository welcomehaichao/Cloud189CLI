package types

import (
	"testing"
)

func TestSessionFields(t *testing.T) {
	session := &Session{
		ResCode:             0,
		ResMessage:          "success",
		LoginName:           "user@example.com",
		SessionKey:          "test_session_key",
		SessionSecret:       "test_session_secret",
		FamilySessionKey:    "family_key",
		FamilySessionSecret: "family_secret",
		AccessToken:         "access_token_123",
		RefreshToken:        "refresh_token_456",
	}

	if session.LoginName != "user@example.com" {
		t.Errorf("LoginName = %s, want user@example.com", session.LoginName)
	}

	if session.SessionKey != "test_session_key" {
		t.Errorf("SessionKey = %s, want test_session_key", session.SessionKey)
	}

	if session.SessionSecret != "test_session_secret" {
		t.Errorf("SessionSecret = %s, want test_session_secret", session.SessionSecret)
	}
}

func TestSessionFamilyFields(t *testing.T) {
	session := &Session{
		FamilySessionKey:    "family_session_key",
		FamilySessionSecret: "family_session_secret",
	}

	if session.FamilySessionKey != "family_session_key" {
		t.Errorf("FamilySessionKey = %s", session.FamilySessionKey)
	}

	if session.FamilySessionSecret != "family_session_secret" {
		t.Errorf("FamilySessionSecret = %s", session.FamilySessionSecret)
	}
}

func TestSessionTokenFields(t *testing.T) {
	session := &Session{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
	}

	if session.AccessToken != "my_access_token" {
		t.Errorf("AccessToken = %s", session.AccessToken)
	}

	if session.RefreshToken != "my_refresh_token" {
		t.Errorf("RefreshToken = %s", session.RefreshToken)
	}
}

func TestAppSession(t *testing.T) {
	appSession := &AppSession{
		Session: Session{
			LoginName:     "app_user",
			SessionKey:    "app_key",
			SessionSecret: "app_secret",
		},
	}

	if appSession.LoginName != "app_user" {
		t.Errorf("AppSession LoginName = %s, want app_user", appSession.LoginName)
	}
}

func TestUserSession(t *testing.T) {
	userSession := &UserSession{
		Session: Session{
			LoginName:     "regular_user",
			SessionKey:    "user_key",
			SessionSecret: "user_secret",
		},
	}

	if userSession.LoginName != "regular_user" {
		t.Errorf("UserSession LoginName = %s, want regular_user", userSession.LoginName)
	}
}

func TestFamilyInfo(t *testing.T) {
	family := &FamilyInfo{
		Count:      5,
		CreateTime: "2024-01-01",
		FamilyID:   12345,
		RemarkName: "MyFamily",
		Type:       1,
		UseFlag:    1,
		UserRole:   1,
	}

	if family.FamilyID != 12345 {
		t.Errorf("FamilyID = %d, want 12345", family.FamilyID)
	}

	if family.RemarkName != "MyFamily" {
		t.Errorf("RemarkName = %s, want MyFamily", family.RemarkName)
	}
}

func TestFamilyInfoList(t *testing.T) {
	list := &FamilyInfoList{
		FamilyInfoResp: []FamilyInfo{
			{FamilyID: 1, RemarkName: "Family1"},
			{FamilyID: 2, RemarkName: "Family2"},
		},
	}

	if len(list.FamilyInfoResp) != 2 {
		t.Errorf("FamilyInfoList length = %d, want 2", len(list.FamilyInfoResp))
	}
}

func TestBaseLoginParam(t *testing.T) {
	param := &BaseLoginParam{
		Lt:           "lt_value",
		ReqId:        "req_id_value",
		ParamId:      "param_id_value",
		CaptchaToken: "captcha_token",
	}

	if param.Lt != "lt_value" {
		t.Errorf("Lt = %s, want lt_value", param.Lt)
	}

	if param.ReqId != "req_id_value" {
		t.Errorf("ReqId = %s, want req_id_value", param.ReqId)
	}
}

func TestQRLoginParam(t *testing.T) {
	qrParam := &QRLoginParam{
		BaseLoginParam: BaseLoginParam{
			Lt:           "lt",
			ReqId:        "req_id",
			ParamId:      "param_id",
			CaptchaToken: "token",
		},
		UUID:       "uuid-123",
		EncodeUUID: "encode-uuid-123",
		EncryUUID:  "encry-uuid-123",
	}

	if qrParam.UUID != "uuid-123" {
		t.Errorf("UUID = %s, want uuid-123", qrParam.UUID)
	}

	if qrParam.Lt != "lt" {
		t.Errorf("Embedded Lt = %s, want lt", qrParam.Lt)
	}
}

func TestSessionKeepAliveFields(t *testing.T) {
	session := &Session{
		KeepAlive:       3600,
		GetFileDiffSpan: 300,
		GetUserInfoSpan: 600,
	}

	if session.KeepAlive != 3600 {
		t.Errorf("KeepAlive = %d, want 3600", session.KeepAlive)
	}
}

func TestSessionResCode(t *testing.T) {
	successSession := &Session{
		ResCode:    0,
		ResMessage: "success",
	}

	failedSession := &Session{
		ResCode:    1,
		ResMessage: "failed",
	}

	if successSession.ResCode != 0 {
		t.Error("Success session should have ResCode 0")
	}

	if failedSession.ResCode == 0 {
		t.Error("Failed session should have non-zero ResCode")
	}
}
