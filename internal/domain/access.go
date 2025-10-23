package domain

import (
	"time"
)

const CollectionNameAccess = "access"

type Access struct {
	Meta
	Name      string         `json:"name" db:"name"`
	Provider  string         `json:"provider" db:"provider"`
	Config    map[string]any `json:"config" db:"config"`
	Reserve   string         `json:"reserve,omitempty" db:"reserve"`
	DeletedAt *time.Time     `json:"deleted" db:"deleted"`
}

type AccessConfigFor1Panel struct {
	ServerUrl                string `json:"serverUrl"`
	ApiVersion               string `json:"apiVersion"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForACMEExternalAccountBinding struct {
	EabKid     string `json:"eabKid,omitempty"`
	EabHmacKey string `json:"eabHmacKey,omitempty"`
}

type AccessConfigForACMECA struct {
	AccessConfigForACMEExternalAccountBinding
	Endpoint string `json:"endpoint"`
}

type AccessConfigForACMEDNS struct {
	ServerUrl   string `json:"serverUrl"`
	Credentials string `json:"credentials"`
}

type AccessConfigForACMEHttpReq struct {
	Endpoint string `json:"endpoint"`
	Mode     string `json:"mode,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AccessConfigForActalisSSL struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForAliyun struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
}

type AccessConfigForAPISIX struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForAWS struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForAzure struct {
	TenantId     string `json:"tenantId"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	CloudName    string `json:"cloudName,omitempty"`
}

type AccessConfigForBaiduCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForBaishan struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForBaotaPanel struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForBaotaPanelGo struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForBaotaWAF struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForBookMyName struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccessConfigForBunny struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForBytePlus struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type AccessConfigForCacheFly struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForCdnfly struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	ApiSecret                string `json:"apiSecret"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForCloudflare struct {
	DnsApiToken  string `json:"dnsApiToken"`
	ZoneApiToken string `json:"zoneApiToken,omitempty"`
}

type AccessConfigForClouDNS struct {
	AuthId       string `json:"authId"`
	AuthPassword string `json:"authPassword"`
}

type AccessConfigForCMCCCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

type AccessConfigForConstellix struct {
	ApiKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
}

type AccessConfigForCTCCCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForDeSEC struct {
	Token string `json:"token"`
}

type AccessConfigForDigitalOcean struct {
	AccessToken string `json:"accessToken"`
}

type AccessConfigForDingTalkBot struct {
	WebhookUrl string `json:"webhookUrl"`
	Secret     string `json:"secret"`
}

type AccessConfigForDiscordBot struct {
	BotToken  string `json:"botToken"`
	ChannelId string `json:"channelId,omitempty"`
}

type AccessConfigForDNSLA struct {
	ApiId     string `json:"apiId"`
	ApiSecret string `json:"apiSecret"`
}

type AccessConfigForDogeCloud struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type AccessConfigForDuckDNS struct {
	Token string `json:"token"`
}

type AccessConfigForDynv6 struct {
	HttpToken string `json:"httpToken"`
}

type AccessConfigForEmail struct {
	SmtpHost        string `json:"smtpHost"`
	SmtpPort        int32  `json:"smtpPort"`
	SmtpTls         bool   `json:"smtpTls"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	SenderAddress   string `json:"senderAddress"`
	SenderName      string `json:"senderName"`
	ReceiverAddress string `json:"receiverAddress,omitempty"`
}

type AccessConfigForFlexCDN struct {
	ServerUrl                string `json:"serverUrl"`
	ApiRole                  string `json:"apiRole"`
	AccessKeyId              string `json:"accessKeyId"`
	AccessKey                string `json:"accessKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForGandinet struct {
	PersonalAccessToken string `json:"personalAccessToken"`
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

type AccessConfigForGname struct {
	AppId  string `json:"appId"`
	AppKey string `json:"appKey"`
}

type AccessConfigForGoDaddy struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
}

type AccessConfigForGoEdge struct {
	ServerUrl                string `json:"serverUrl"`
	ApiRole                  string `json:"apiRole"`
	AccessKeyId              string `json:"accessKeyId"`
	AccessKey                string `json:"accessKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForGoogleTrustServices struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForHetzner struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForHostinger struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForHuaweiCloud struct {
	AccessKeyId         string `json:"accessKeyId"`
	SecretAccessKey     string `json:"secretAccessKey"`
	EnterpriseProjectId string `json:"enterpriseProjectId,omitempty"`
}

type AccessConfigForIONOS struct {
	ApiKeyPublicPrefix string `json:"apiKeyPublicPrefix"`
	ApiKeySecret       string `json:"apiKeySecret"`
}

type AccessConfigForJDCloud struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

type AccessConfigForKong struct {
	ServerUrl                string `json:"serverUrl"`
	ApiToken                 string `json:"apiToken,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForKubernetes struct {
	KubeConfig string `json:"kubeConfig,omitempty"`
}

type AccessConfigForLarkBot struct {
	WebhookUrl string `json:"webhookUrl"`
}

type AccessConfigForLeCDN struct {
	ServerUrl                string `json:"serverUrl"`
	ApiVersion               string `json:"apiVersion"`
	ApiRole                  string `json:"apiRole"`
	Username                 string `json:"username"`
	Password                 string `json:"password"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForLinode struct {
	AccessToken string `json:"accessToken"`
}

type AccessConfigForMattermost struct {
	ServerUrl string `json:"serverUrl"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	ChannelId string `json:"channelId,omitempty"`
}

type AccessConfigForNamecheap struct {
	Username string `json:"username"`
	ApiKey   string `json:"apiKey"`
}

type AccessConfigForNameDotCom struct {
	Username string `json:"username"`
	ApiToken string `json:"apiToken"`
}

type AccessConfigForNameSilo struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForNetcup struct {
	CustomerNumber string `json:"customerNumber"`
	ApiKey         string `json:"apiKey"`
	ApiPassword    string `json:"apiPassword"`
}

type AccessConfigForNetlify struct {
	ApiToken string `json:"apiToken"`
}

type AccessConfigForNS1 struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForPorkbun struct {
	ApiKey       string `json:"apiKey"`
	SecretApiKey string `json:"secretApiKey"`
}

type AccessConfigForPowerDNS struct {
	ServerUrl                string `json:"serverUrl"`
	ApiKey                   string `json:"apiKey"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForProxmoxVE struct {
	ServerUrl                string `json:"serverUrl"`
	ApiToken                 string `json:"apiToken"`
	ApiTokenSecret           string `json:"apiTokenSecret,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForQiniu struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type AccessConfigForRainYun struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForRatPanel struct {
	ServerUrl                string `json:"serverUrl"`
	AccessTokenId            int32  `json:"accessTokenId"`
	AccessToken              string `json:"accessToken"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForRFC2136 struct {
	Host          string `json:"host"`
	Port          int32  `json:"port"`
	TsigAlgorithm string `json:"tsigAlgorithm,omitempty"`
	TsigKey       string `json:"tsigKey,omitempty"`
	TsigSecret    string `json:"tsigSecret,omitempty"`
}

type AccessConfigForSafeLine struct {
	ServerUrl                string `json:"serverUrl"`
	ApiToken                 string `json:"apiToken"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForSlackBot struct {
	BotToken  string `json:"botToken"`
	ChannelId string `json:"channelId,omitempty"`
}

type AccessConfigForSpaceship struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
}

type AccessConfigForSSH struct {
	Host          string `json:"host"`
	Port          int32  `json:"port"`
	AuthMethod    string `json:"authMethod,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Key           string `json:"key,omitempty"`
	KeyPassphrase string `json:"keyPassphrase,omitempty"`
	JumpServers   []struct {
		Host          string `json:"host"`
		Port          int32  `json:"port"`
		AuthMethod    string `json:"authMethod,omitempty"`
		Username      string `json:"username,omitempty"`
		Password      string `json:"password,omitempty"`
		Key           string `json:"key,omitempty"`
		KeyPassphrase string `json:"keyPassphrase,omitempty"`
	} `json:"jumpServers,omitempty"`
}

type AccessConfigForSSLCom struct {
	AccessConfigForACMEExternalAccountBinding
}

type AccessConfigForTechnitiumDNS struct {
	ServerUrl                string `json:"serverUrl"`
	ApiToken                 string `json:"apiToken"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForTelegramBot struct {
	BotToken string `json:"botToken"`
	ChatId   int64  `json:"chatId,omitempty"`
}

type AccessConfigForTencentCloud struct {
	SecretId  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
}

type AccessConfigForUCloud struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	ProjectId  string `json:"projectId,omitempty"`
}

type AccessConfigForUniCloud struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccessConfigForUpyun struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccessConfigForVercel struct {
	ApiAccessToken string `json:"apiAccessToken"`
	TeamId         string `json:"teamId,omitempty"`
}

type AccessConfigForVolcEngine struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type AccessConfigForVultr struct {
	ApiKey string `json:"apiKey"`
}

type AccessConfigForWangsu struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	ApiKey          string `json:"apiKey"`
}

type AccessConfigForWebhook struct {
	Url                      string `json:"url"`
	Method                   string `json:"method,omitempty"`
	HeadersString            string `json:"headers,omitempty"`
	DataString               string `json:"data,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

type AccessConfigForWeComBot struct {
	WebhookUrl string `json:"webhookUrl"`
}

type AccessConfigForWestcn struct {
	Username    string `json:"username"`
	ApiPassword string `json:"apiPassword"`
}

type AccessConfigForZeroSSL struct {
	AccessConfigForACMEExternalAccountBinding
}
