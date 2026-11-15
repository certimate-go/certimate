package dtos

type SetupStatusResp struct {
	Initialized bool `json:"initialized"`
}

type SetupReq struct {
	Token           string `json:"token"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}
