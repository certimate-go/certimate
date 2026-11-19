package domain

import (
	"time"
)

const CollectionNameAccess = "access"

type Access struct {
	Meta
	Name      string         `db:"name"     json:"name"`
	Provider  string         `db:"provider" json:"provider"`
	Config    map[string]any `db:"config"   json:"config"`
	Reserve   string         `db:"reserve"  json:"reserve,omitempty"`
	DeletedAt *time.Time     `db:"deleted" json:"deleted"`
}

type AccessConfigForACMEExternalAccountBinding struct {
	EabKid     string `json:"eabKid,omitempty"`
	EabHmacKey string `json:"eabHmacKey,omitempty"`
}

type AccessConfigForACMECA struct {
	AccessConfigForACMEExternalAccountBinding
	Endpoint string `json:"endpoint"`
}

type AccessConfigForActalisSSL struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForAliyun struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
}

type AccessConfigForAWS struct {
	AuthMethod      string `json:"authMethod"`
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForAzure struct {
	TenantId          string `json:"tenantId"`
	ClientId          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	SubscriptionId    string `json:"subscriptionId,omitempty"`
	ResourceGroupName string `json:"resourceGroupName,omitempty"`
	CloudName         string `json:"cloudName,omitempty"`
}

type AccessConfigForBaiduCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForBunny struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForBytePlus struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	ProjectName     string `json:"projectName,omitempty"`
}

type AccessConfigForCloudflare struct {
	ApiToken        string `json:"apiToken"`
	ApiTokenForZone string `json:"apiTokenForZone,omitempty"`
}

type AccessConfigForCMCCCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

type AccessConfigForCPanel struct {
	ServerUrl                string `json:"serverUrl"`
	Username                 string `json:"username"`
	ApiToken                 string `json:"apiToken"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForCTCCCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForDigitalOcean struct {
	AccessToken string `json:"accessToken"`
}

type AccessConfigForEmail struct {
	SmtpHost                 string `json:"smtpHost"`
	SmtpPort                 int32  `json:"smtpPort"`
	SmtpTls                  bool   `json:"smtpTls"`
	Username                 string `json:"username"`
	Password                 string `json:"password"`
	SenderAddress            string `json:"senderAddress"`
	SenderName               string `json:"senderName"`
	ReceiverAddress          string `json:"receiverAddress,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForFTP struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AccessConfigForGcore struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForGlobalSectigo struct {
	AccessConfigForACMEExternalAccountBinding
	ValidationType string `json:"validationType"`
}

type AccessConfigForGlobalSignAtlas struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForGoogleCloud struct {
	ProjectId         string `json:"projectId"`
	ServiceAccountKey string `json:"serviceAccountKey"`
}

type AccessConfigForGoogleTrustServices struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForHuaweiCloud struct {
	AccessKeyId         string `json:"accessKeyId"`
	SecretAccessKey     string `json:"secretAccessKey"`
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
}

type AccessConfigForJDCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

type AccessConfigForLinode struct {
	AccessToken string `json:"accessToken"`
}

type AccessConfigForLiteSSL struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForNetlify struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForOracleCloud struct {
	AuthMethod           string `json:"authMethod"`
	PrivateKey           string `json:"privateKey,omitempty"`
	PrivateKeyPassphrase string `json:"privateKeyPassphrase,omitempty"`
	PublicKeyFingerprint string `json:"publicKeyFingerprint,omitempty"`
	TenancyOcid          string `json:"tenancyOcid,omitempty"`
	UserOcid             string `json:"userOcid,omitempty"`
}

type AccessConfigForPowerDNS struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForQingCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForRainYun struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForRFC2136 struct {
	Host            string `json:"host"`
	Port            int32  `json:"port"`
	TsigAlgorithm   string `json:"tsigAlgorithm,omitempty"`
	TsigKey         string `json:"tsigKey,omitempty"`
	TsigSecret      string `json:"tsigSecret,omitempty"`
	TsigGssRealm    string `json:"tsigGssRealm,omitempty"`
	TsigGssUsername string `json:"tsigGssUsername,omitempty"`
	TsigGssPassword string `json:"tsigGssPassword,omitempty"`
}

type AccessConfigForS3 struct {
	Endpoint                 string `json:"endpoint"`
	AccessKey                string `json:"accessKey"`
	SecretKey                string `json:"secretKey"`
	SignatureVersion         string `json:"signatureVersion,omitempty"`
	UsePathStyle             bool   `json:"usePathStyle,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForSSH struct {
	Host           string `json:"host"`
	Port           int32  `json:"port"`
	AuthMethod     string `json:"authMethod"`
	Username       string `json:"username"`
	Password       string `json:"password,omitempty"`
	Key            string `json:"key,omitempty"`
	KeyPassphrase  string `json:"keyPassphrase,omitempty"`
	KeyCertificate string `json:"keyCertificate,omitempty"`
	JumpServers    []struct {
		Host           string `json:"host"`
		Port           int32  `json:"port"`
		AuthMethod     string `json:"authMethod"`
		Username       string `json:"username"`
		Password       string `json:"password,omitempty"`
		Key            string `json:"key,omitempty"`
		KeyPassphrase  string `json:"keyPassphrase,omitempty"`
		KeyCertificate string `json:"keyCertificate,omitempty"`
	} `json:"jumpServers,omitempty"`
}

type AccessConfigForSSLCom struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForTencentCloud struct {
	SecretId  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
	ProjectId int64  `json:"projectId,omitempty"`
}

type AccessConfigForUCloud struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
	ProjectId  string `json:"projectId,omitempty"`
}

type AccessConfigForVercel struct {
	ApiAccessToken string `json:"apiAccessToken"`
	TeamId         string `json:"teamId,omitempty"`
}

type AccessConfigForVolcEngine struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	ProjectName     string `json:"projectName,omitempty"`
}

type AccessConfigForWebhook struct {
	Url                      string `json:"url"`
	Method                   string `json:"method,omitempty"`
	HeadersString            string `json:"headers,omitempty"`
	DataString               string `json:"data,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForYandexCloud struct {
	FolderId          string `json:"folderId"`
	ServiceAccountKey string `json:"serviceAccountKey"`
}

type AccessConfigForZeroSSL struct {
	AccessConfigForACMEExternalAccountBinding
}
