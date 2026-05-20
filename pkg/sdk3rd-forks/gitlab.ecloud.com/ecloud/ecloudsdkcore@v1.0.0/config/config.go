package config

type Config struct {
	AccessKey      string `json:"accessKey,string"`
	SecretKey      string `json:"secretKey,string"`
	PoolId         string `json:"poolId,string"`
	ReadTimeOut    int    `json:"readTimeOut,int"`
	ConnectTimeout int    `json:"connectTimeout,int"`
}
