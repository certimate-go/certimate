package dtos

type MatrixTestSendReq struct {
	Config   map[string]any `json:"config"`
	AccessId string         `json:"accessId,omitempty"`
	Subject  string         `json:"subject,omitempty"`
	Message  string         `json:"message,omitempty"`
}

type MatrixTestSendResp struct {
	Ok                 bool   `json:"ok"`
	UserId             string `json:"userId,omitempty"`
	SessionAccessToken string `json:"sessionAccessToken,omitempty"`
	SessionDeviceId    string `json:"sessionDeviceId,omitempty"`
	SessionSaved       bool   `json:"sessionSaved,omitempty"`
}
