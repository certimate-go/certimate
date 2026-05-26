package dtos

type MatrixVerifyConnectionReq struct {
	Config   map[string]any `json:"config"`
	AccessId string         `json:"accessId,omitempty"`
}

type MatrixVerifyConnectionResp struct {
	Ok                 bool                  `json:"ok"`
	UserId             string                `json:"userId,omitempty"`
	SessionAccessToken string                `json:"sessionAccessToken,omitempty"`
	SessionDeviceId    string                `json:"sessionDeviceId,omitempty"`
	SessionSaved       bool                  `json:"sessionSaved,omitempty"`
	Steps              []MatrixVerifyStepDTO `json:"steps"`
}

type MatrixVerifyStepDTO struct {
	Name          string `json:"name"`
	Ok            bool   `json:"ok"`
	Message       string `json:"message"`
	Detail        string `json:"detail,omitempty"`
	Code          string `json:"code,omitempty"`
	RetryAfterSec int    `json:"retryAfterSec,omitempty"`
}
