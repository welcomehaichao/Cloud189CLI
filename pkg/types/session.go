package types

type Session struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`

	LoginName string `json:"loginName"`

	KeepAlive       int `json:"keepAlive"`
	GetFileDiffSpan int `json:"getFileDiffSpan"`
	GetUserInfoSpan int `json:"getUserInfoSpan"`

	SessionKey    string `json:"sessionKey"`
	SessionSecret string `json:"sessionSecret"`

	FamilySessionKey    string `json:"familySessionKey"`
	FamilySessionSecret string `json:"familySessionSecret"`

	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`

	IsSaveName string `json:"isSaveName"`
}

type AppSession struct {
	Session
}

type UserSession struct {
	Session
}

type FamilyInfo struct {
	Count      int    `json:"count"`
	CreateTime string `json:"createTime"`
	FamilyID   int64  `json:"familyId"`
	RemarkName string `json:"remarkName"`
	Type       int    `json:"type"`
	UseFlag    int    `json:"useFlag"`
	UserRole   int    `json:"userRole"`
}

type FamilyInfoList struct {
	FamilyInfoResp []FamilyInfo `json:"familyInfoResp"`
}

type BaseLoginParam struct {
	Lt           string
	ReqId        string
	ParamId      string
	CaptchaToken string
}

type QRLoginParam struct {
	BaseLoginParam
	UUID       string `json:"uuid"`
	EncodeUUID string `json:"encodeuuid"`
	EncryUUID  string `json:"encryuuid"`
}
