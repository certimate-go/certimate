package dtos

type InstallerStatusResp struct {
	Installed bool `json:"installed"`
}

type InstallerPrepareResp struct {
	Installed      bool   `json:"installed"`
	Token          string `json:"token,omitempty"`
	TokenExpiredAt int64  `json:"expired_at,omitempty"`
}

type InstallerInitializeSuperuserReq struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type InstallerInitializeSuperuserResp struct{}
