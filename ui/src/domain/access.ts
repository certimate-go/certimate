export interface AccessModel extends BaseModel {
  name: string;
  provider: string;
  config: /*
    注意：如果追加新的类型，请保持以 ASCII 排序。
    NOTICE: If you add new type, please keep ASCII order.
  */ Record<string, unknown> &
    (
      | AccessConfigFor1Panel
      | AccessConfigForACMECA
      | AccessConfigForACMEHttpReq
      | AccessConfigForAliyun
      | AccessConfigForAPISIX
      | AccessConfigForAWS
      | AccessConfigForAzure
      | AccessConfigForBaiduCloud
      | AccessConfigForBaishan
      | AccessConfigForBaotaPanel
      | AccessConfigForBaotaWAF
      | AccessConfigForBunny
      | AccessConfigForBytePlus
      | AccessConfigForCacheFly
      | AccessConfigForCdnfly
      | AccessConfigForCloudflare
      | AccessConfigForClouDNS
      | AccessConfigForCMCCCloud
      | AccessConfigForConstellix
      | AccessConfigForCTCCCloud
      | AccessConfigForDeSEC
      | AccessConfigForDigitalOcean
      | AccessConfigForDingTalkBot
      | AccessConfigForDiscordBot
      | AccessConfigForDNSLA
      | AccessConfigForDogeCloud
      | AccessConfigForDuckDNS
      | AccessConfigForDynv6
      | AccessConfigForEdgio
      | AccessConfigForEmail
      | AccessConfigForFlexCDN
      | AccessConfigForGcore
      | AccessConfigForGname
      | AccessConfigForGoDaddy
      | AccessConfigForGoEdge
      | AccessConfigForGoogleTrustServices
      | AccessConfigForHetzner
      | AccessConfigForHuaweiCloud
      | AccessConfigForJDCloud
      | AccessConfigForKong
      | AccessConfigForKubernetes
      | AccessConfigForLarkBot
      | AccessConfigForLeCDN
      | AccessConfigForMattermost
      | AccessConfigForNamecheap
      | AccessConfigForNameDotCom
      | AccessConfigForNameSilo
      | AccessConfigForNetcup
      | AccessConfigForNetlify
      | AccessConfigForPorkbun
      | AccessConfigForPowerDNS
      | AccessConfigForProxmoxVE
      | AccessConfigForQiniu
      | AccessConfigForRainYun
      | AccessConfigForRatPanel
      | AccessConfigForSafeLine
      | AccessConfigForSlackBot
      | AccessConfigForSpaceship
      | AccessConfigForSSH
      | AccessConfigForSSLCom
      | AccessConfigForTelegramBot
      | AccessConfigForTencentCloud
      | AccessConfigForUCloud
      | AccessConfigForUniCloud
      | AccessConfigForUpyun
      | AccessConfigForVercel
      | AccessConfigForVolcEngine
      | AccessConfigForWangsu
      | AccessConfigForWebhook
      | AccessConfigForWeComBot
      | AccessConfigForWestcn
      | AccessConfigForZeroSSL
    );
  reserve?: "ca" | "notification";
}

// #region AccessConfig
export type AccessConfigFor1Panel = {
  serverUrl: string;
  apiVersion: string;
  apiKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForACMECA = {
  endpoint: string;
  eabKid?: string;
  eabHmacKey?: string;
};

export type AccessConfigForACMEHttpReq = {
  endpoint: string;
  mode?: string;
  username?: string;
  password?: string;
};

export type AccessConfigForAliyun = {
  accessKeyId: string;
  accessKeySecret: string;
  resourceGroupId?: string;
};

export type AccessConfigForAPISIX = {
  serverUrl: string;
  apiKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForAWS = {
  accessKeyId: string;
  secretAccessKey: string;
};

export type AccessConfigForAzure = {
  tenantId: string;
  clientId: string;
  clientSecret: string;
  environment?: string;
};

export type AccessConfigForBaiduCloud = {
  accessKeyId: string;
  secretAccessKey: string;
};

export type AccessConfigForBaishan = {
  apiToken: string;
};

export type AccessConfigForBaotaPanel = {
  serverUrl: string;
  apiKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForBaotaWAF = {
  serverUrl: string;
  apiKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForBunny = {
  apiKey: string;
};

export type AccessConfigForBytePlus = {
  accessKey: string;
  secretKey: string;
};

export type AccessConfigForCacheFly = {
  apiToken: string;
};

export type AccessConfigForCdnfly = {
  serverUrl: string;
  apiKey: string;
  apiSecret: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForCloudflare = {
  dnsApiToken: string;
  zoneApiToken?: string;
};

export type AccessConfigForClouDNS = {
  authId: string;
  authPassword: string;
};

export type AccessConfigForCMCCCloud = {
  accessKeyId: string;
  accessKeySecret: string;
};

export type AccessConfigForConstellix = {
  apiKey: string;
  secretKey: string;
};

export type AccessConfigForCTCCCloud = {
  accessKeyId: string;
  secretAccessKey: string;
};

export type AccessConfigForDeSEC = {
  token: string;
};

export type AccessConfigForDigitalOcean = {
  accessToken: string;
};

export type AccessConfigForDingTalkBot = {
  webhookUrl: string;
  secret?: string;
};

export type AccessConfigForDiscordBot = {
  botToken: string;
  defaultChannelId?: string;
};

export type AccessConfigForDNSLA = {
  apiId: string;
  apiSecret: string;
};

export type AccessConfigForDogeCloud = {
  accessKey: string;
  secretKey: string;
};

export type AccessConfigForDuckDNS = {
  token: string;
};

export type AccessConfigForDynv6 = {
  httpToken: string;
};

export type AccessConfigForEdgio = {
  clientId: string;
  clientSecret: string;
};

export type AccessConfigForEmail = {
  smtpHost: string;
  smtpPort: number;
  smtpTls: boolean;
  username: string;
  password: string;
  defaultSenderAddress?: string;
  defaultReceiverAddress?: string;
};

export type AccessConfigForFlexCDN = {
  serverUrl: string;
  apiRole: string;
  accessKeyId: string;
  accessKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForGcore = {
  apiToken: string;
};

export type AccessConfigForGname = {
  appId: string;
  appKey: string;
};

export type AccessConfigForGoDaddy = {
  apiKey: string;
  apiSecret: string;
};

export type AccessConfigForGoEdge = {
  serverUrl: string;
  apiRole: string;
  accessKeyId: string;
  accessKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForGoogleTrustServices = {
  eabKid: string;
  eabHmacKey: string;
};

export type AccessConfigForHetzner = {
  apiToken: string;
};

export type AccessConfigForHuaweiCloud = {
  accessKeyId: string;
  secretAccessKey: string;
  enterpriseProjectId?: string;
};

export type AccessConfigForJDCloud = {
  accessKeyId: string;
  accessKeySecret: string;
};

export type AccessConfigForKong = {
  serverUrl: string;
  apiToken: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForKubernetes = {
  kubeConfig?: string;
};

export type AccessConfigForLarkBot = {
  webhookUrl: string;
};

export type AccessConfigForLeCDN = {
  serverUrl: string;
  apiVersion: string;
  apiRole: string;
  username: string;
  password: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForMattermost = {
  serverUrl: string;
  username: string;
  password: string;
  defaultChannelId?: string;
};

export type AccessConfigForNamecheap = {
  username: string;
  apiKey: string;
};

export type AccessConfigForNameDotCom = {
  username: string;
  apiToken: string;
};

export type AccessConfigForNameSilo = {
  apiKey: string;
};

export type AccessConfigForNetcup = {
  customerNumber: string;
  apiKey: string;
  apiPassword: string;
};

export type AccessConfigForNetlify = {
  apiToken: string;
};

export type AccessConfigForNS1 = {
  apiKey: string;
};

export type AccessConfigForPorkbun = {
  apiKey: string;
  secretApiKey: string;
};

export type AccessConfigForPowerDNS = {
  serverUrl: string;
  apiKey: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForProxmoxVE = {
  serverUrl: string;
  apiToken: string;
  apiTokenSecret?: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForQiniu = {
  accessKey: string;
  secretKey: string;
};

export type AccessConfigForRainYun = {
  apiKey: string;
};

export type AccessConfigForRatPanel = {
  serverUrl: string;
  accessTokenId: number;
  accessToken: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForSafeLine = {
  serverUrl: string;
  apiToken: string;
  allowInsecureConnections?: boolean;
};

export type AccessConfigForSlackBot = {
  botToken: string;
  defaultChannelId?: string;
};

export type AccessConfigForSpaceship = {
  apiKey: string;
  apiSecret: string;
};

export type AccessConfigForSSH = {
  host: string;
  port: number;
  authMethod?: string;
  username?: string;
  password?: string;
  key?: string;
  keyPassphrase?: string;
};

export type AccessConfigForSSLCom = {
  eabKid: string;
  eabHmacKey: string;
};

export type AccessConfigForTelegramBot = {
  botToken: string;
  defaultChatId?: number;
};

export type AccessConfigForTencentCloud = {
  secretId: string;
  secretKey: string;
};

export type AccessConfigForUCloud = {
  privateKey: string;
  publicKey: string;
  projectId?: string;
};

export type AccessConfigForUniCloud = {
  username: string;
  password: string;
};

export type AccessConfigForUpyun = {
  username: string;
  password: string;
};

export type AccessConfigForVercel = {
  apiAccessToken: string;
  teamId?: string;
};

export type AccessConfigForVolcEngine = {
  accessKeyId: string;
  secretAccessKey: string;
};

export type AccessConfigForWangsu = {
  accessKeyId: string;
  accessKeySecret: string;
  apiKey: string;
};

export type AccessConfigForWebhook = {
  url: string;
  method: string;
  headers?: string;
  allowInsecureConnections?: boolean;
  defaultDataForDeployment?: string;
  defaultDataForNotification?: string;
};

export type AccessConfigForWeComBot = {
  webhookUrl: string;
};

export type AccessConfigForWestcn = {
  username: string;
  apiPassword: string;
};

export type AccessConfigForZeroSSL = {
  eabKid: string;
  eabHmacKey: string;
};
// #endregion
