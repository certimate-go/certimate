import { useEffect, useState } from "react";

import { DEPLOYMENT_PROVIDERS, type DeploymentProviderType } from "@/domain/provider";

import BizDeployNodeConfigFieldsProvider1Panel from "./BizDeployNodeConfigFieldsProvider1Panel.tsx";
import BizDeployNodeConfigFieldsProvider1PanelConsole from "./BizDeployNodeConfigFieldsProvider1PanelConsole";
import BizDeployNodeConfigFieldsProviderAliyunALB from "./BizDeployNodeConfigFieldsProviderAliyunALB";
import BizDeployNodeConfigFieldsProviderAliyunAPIGW from "./BizDeployNodeConfigFieldsProviderAliyunAPIGW";
import BizDeployNodeConfigFieldsProviderAliyunCAS from "./BizDeployNodeConfigFieldsProviderAliyunCAS";
import BizDeployNodeConfigFieldsProviderAliyunCASDeploy from "./BizDeployNodeConfigFieldsProviderAliyunCASDeploy";
import BizDeployNodeConfigFieldsProviderAliyunCDN from "./BizDeployNodeConfigFieldsProviderAliyunCDN";
import BizDeployNodeConfigFieldsProviderAliyunCLB from "./BizDeployNodeConfigFieldsProviderAliyunCLB";
import BizDeployNodeConfigFieldsProviderAliyunDCDN from "./BizDeployNodeConfigFieldsProviderAliyunDCDN";
import BizDeployNodeConfigFieldsProviderAliyunDDoSPro from "./BizDeployNodeConfigFieldsProviderAliyunDDoSPro";
import BizDeployNodeConfigFieldsProviderAliyunESA from "./BizDeployNodeConfigFieldsProviderAliyunESA";
import BizDeployNodeConfigFieldsProviderAliyunFC from "./BizDeployNodeConfigFieldsProviderAliyunFC";
import BizDeployNodeConfigFieldsProviderAliyunGA from "./BizDeployNodeConfigFieldsProviderAliyunGA";
import BizDeployNodeConfigFieldsProviderAliyunLive from "./BizDeployNodeConfigFieldsProviderAliyunLive";
import BizDeployNodeConfigFieldsProviderAliyunNLB from "./BizDeployNodeConfigFieldsProviderAliyunNLB";
import BizDeployNodeConfigFieldsProviderAliyunOSS from "./BizDeployNodeConfigFieldsProviderAliyunOSS";
import BizDeployNodeConfigFieldsProviderAliyunVOD from "./BizDeployNodeConfigFieldsProviderAliyunVOD";
import BizDeployNodeConfigFieldsProviderAliyunWAF from "./BizDeployNodeConfigFieldsProviderAliyunWAF";
import BizDeployNodeConfigFieldsProviderAPISIX from "./BizDeployNodeConfigFieldsProviderAPISIX";
import BizDeployNodeConfigFieldsProviderAWSACM from "./BizDeployNodeConfigFieldsProviderAWSACM";
import BizDeployNodeConfigFieldsProviderAWSCloudFront from "./BizDeployNodeConfigFieldsProviderAWSCloudFront";
import BizDeployNodeConfigFieldsProviderAWSIAM from "./BizDeployNodeConfigFieldsProviderAWSIAM";
import BizDeployNodeConfigFieldsProviderAzureKeyVault from "./BizDeployNodeConfigFieldsProviderAzureKeyVault";
import BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB from "./BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB";
import BizDeployNodeConfigFieldsProviderBaiduCloudBLB from "./BizDeployNodeConfigFieldsProviderBaiduCloudBLB";
import BizDeployNodeConfigFieldsProviderBaiduCloudCDN from "./BizDeployNodeConfigFieldsProviderBaiduCloudCDN";
import BizDeployNodeConfigFieldsProviderBaishanCDN from "./BizDeployNodeConfigFieldsProviderBaishanCDN";
import BizDeployNodeConfigFieldsProviderBaotaPanel from "./BizDeployNodeConfigFieldsProviderBaotaPanel.tsx";
import BizDeployNodeConfigFieldsProviderBaotaPanelConsole from "./BizDeployNodeConfigFieldsProviderBaotaPanelConsole";
import BizDeployNodeConfigFieldsProviderBaotaPanelGo from "./BizDeployNodeConfigFieldsProviderBaotaPanelGo.tsx";
import BizDeployNodeConfigFieldsProviderBaotaPanelGoConsole from "./BizDeployNodeConfigFieldsProviderBaotaPanelGoConsole";
import BizDeployNodeConfigFieldsProviderBaotaWAF from "./BizDeployNodeConfigFieldsProviderBaotaWAF.tsx";
import BizDeployNodeConfigFieldsProviderBaotaWAFConsole from "./BizDeployNodeConfigFieldsProviderBaotaWAFConsole.tsx";
import BizDeployNodeConfigFieldsProviderBunnyCDN from "./BizDeployNodeConfigFieldsProviderBunnyCDN";
import BizDeployNodeConfigFieldsProviderBytePlusCDN from "./BizDeployNodeConfigFieldsProviderBytePlusCDN";
import BizDeployNodeConfigFieldsProviderCdnfly from "./BizDeployNodeConfigFieldsProviderCdnfly";
import BizDeployNodeConfigFieldsProviderCPanel from "./BizDeployNodeConfigFieldsProviderCPanel.tsx";
import BizDeployNodeConfigFieldsProviderCTCCCloudAO from "./BizDeployNodeConfigFieldsProviderCTCCCloudAO";
import BizDeployNodeConfigFieldsProviderCTCCCloudCDN from "./BizDeployNodeConfigFieldsProviderCTCCCloudCDN";
import BizDeployNodeConfigFieldsProviderCTCCCloudELB from "./BizDeployNodeConfigFieldsProviderCTCCCloudELB";
import BizDeployNodeConfigFieldsProviderCTCCCloudFaaS from "./BizDeployNodeConfigFieldsProviderCTCCCloudFaaS";
import BizDeployNodeConfigFieldsProviderCTCCCloudICDN from "./BizDeployNodeConfigFieldsProviderCTCCCloudICDN";
import BizDeployNodeConfigFieldsProviderCTCCCloudLVDN from "./BizDeployNodeConfigFieldsProviderCTCCCloudLVDN";
import BizDeployNodeConfigFieldsProviderDogeCloudCDN from "./BizDeployNodeConfigFieldsProviderDogeCloudCDN";
import BizDeployNodeConfigFieldsProviderFlexCDN from "./BizDeployNodeConfigFieldsProviderFlexCDN";
import BizDeployNodeConfigFieldsProviderGcoreCDN from "./BizDeployNodeConfigFieldsProviderGcoreCDN";
import BizDeployNodeConfigFieldsProviderGoEdge from "./BizDeployNodeConfigFieldsProviderGoEdge";
import BizDeployNodeConfigFieldsProviderHuaweiCloudCDN from "./BizDeployNodeConfigFieldsProviderHuaweiCloudCDN";
import BizDeployNodeConfigFieldsProviderHuaweiCloudELB from "./BizDeployNodeConfigFieldsProviderHuaweiCloudELB";
import BizDeployNodeConfigFieldsProviderHuaweiCloudOBS from "./BizDeployNodeConfigFieldsProviderHuaweiCloudOBS";
import BizDeployNodeConfigFieldsProviderHuaweiCloudWAF from "./BizDeployNodeConfigFieldsProviderHuaweiCloudWAF";
import BizDeployNodeConfigFieldsProviderJDCloudALB from "./BizDeployNodeConfigFieldsProviderJDCloudALB";
import BizDeployNodeConfigFieldsProviderJDCloudCDN from "./BizDeployNodeConfigFieldsProviderJDCloudCDN";
import BizDeployNodeConfigFieldsProviderJDCloudLive from "./BizDeployNodeConfigFieldsProviderJDCloudLive";
import BizDeployNodeConfigFieldsProviderJDCloudVOD from "./BizDeployNodeConfigFieldsProviderJDCloudVOD";
import BizDeployNodeConfigFieldsProviderKong from "./BizDeployNodeConfigFieldsProviderKong";
import BizDeployNodeConfigFieldsProviderKsyunCDN from "./BizDeployNodeConfigFieldsProviderKsyunCDN";
import BizDeployNodeConfigFieldsProviderKubernetesSecret from "./BizDeployNodeConfigFieldsProviderKubernetesSecret";
import BizDeployNodeConfigFieldsProviderLeCDN from "./BizDeployNodeConfigFieldsProviderLeCDN";
import BizDeployNodeConfigFieldsProviderLocal from "./BizDeployNodeConfigFieldsProviderLocal";
import BizDeployNodeConfigFieldsProviderMohuaMVH from "./BizDeployNodeConfigFieldsProviderMohuaMVH";
import BizDeployNodeConfigFieldsProviderNetlify from "./BizDeployNodeConfigFieldsProviderNetlify.tsx";
import BizDeployNodeConfigFieldsProviderNginxProxyManager from "./BizDeployNodeConfigFieldsProviderNginxProxyManager";
import BizDeployNodeConfigFieldsProviderProxmoxVE from "./BizDeployNodeConfigFieldsProviderProxmoxVE";
import BizDeployNodeConfigFieldsProviderQiniuCDN from "./BizDeployNodeConfigFieldsProviderQiniuCDN";
import BizDeployNodeConfigFieldsProviderQiniuKodo from "./BizDeployNodeConfigFieldsProviderQiniuKodo";
import BizDeployNodeConfigFieldsProviderQiniuPili from "./BizDeployNodeConfigFieldsProviderQiniuPili";
import BizDeployNodeConfigFieldsProviderRainYunRCDN from "./BizDeployNodeConfigFieldsProviderRainYunRCDN";
import BizDeployNodeConfigFieldsProviderRatPanel from "./BizDeployNodeConfigFieldsProviderRatPanel.tsx";
import BizDeployNodeConfigFieldsProviderS3 from "./BizDeployNodeConfigFieldsProviderS3";
import BizDeployNodeConfigFieldsProviderSafeLine from "./BizDeployNodeConfigFieldsProviderSafeLine.tsx";
import BizDeployNodeConfigFieldsProviderSSH from "./BizDeployNodeConfigFieldsProviderSSH";
import BizDeployNodeConfigFieldsProviderSynologyDSM from "./BizDeployNodeConfigFieldsProviderSynologyDSM";
import BizDeployNodeConfigFieldsProviderTencentCloudCDN from "./BizDeployNodeConfigFieldsProviderTencentCloudCDN";
import BizDeployNodeConfigFieldsProviderTencentCloudCLB from "./BizDeployNodeConfigFieldsProviderTencentCloudCLB";
import BizDeployNodeConfigFieldsProviderTencentCloudCOS from "./BizDeployNodeConfigFieldsProviderTencentCloudCOS";
import BizDeployNodeConfigFieldsProviderTencentCloudCSS from "./BizDeployNodeConfigFieldsProviderTencentCloudCSS";
import BizDeployNodeConfigFieldsProviderTencentCloudECDN from "./BizDeployNodeConfigFieldsProviderTencentCloudECDN";
import BizDeployNodeConfigFieldsProviderTencentCloudEO from "./BizDeployNodeConfigFieldsProviderTencentCloudEO";
import BizDeployNodeConfigFieldsProviderTencentCloudGAAP from "./BizDeployNodeConfigFieldsProviderTencentCloudGAAP";
import BizDeployNodeConfigFieldsProviderTencentCloudSCF from "./BizDeployNodeConfigFieldsProviderTencentCloudSCF";
import BizDeployNodeConfigFieldsProviderTencentCloudSSL from "./BizDeployNodeConfigFieldsProviderTencentCloudSSL";
import BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy from "./BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy";
import BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate from "./BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate";
import BizDeployNodeConfigFieldsProviderTencentCloudVOD from "./BizDeployNodeConfigFieldsProviderTencentCloudVOD";
import BizDeployNodeConfigFieldsProviderTencentCloudWAF from "./BizDeployNodeConfigFieldsProviderTencentCloudWAF";
import BizDeployNodeConfigFieldsProviderUCloudUALB from "./BizDeployNodeConfigFieldsProviderUCloudUALB";
import BizDeployNodeConfigFieldsProviderUCloudUCDN from "./BizDeployNodeConfigFieldsProviderUCloudUCDN";
import BizDeployNodeConfigFieldsProviderUCloudUCLB from "./BizDeployNodeConfigFieldsProviderUCloudUCLB";
import BizDeployNodeConfigFieldsProviderUCloudUEWAF from "./BizDeployNodeConfigFieldsProviderUCloudUEWAF";
import BizDeployNodeConfigFieldsProviderUCloudUPathX from "./BizDeployNodeConfigFieldsProviderUCloudUPathX";
import BizDeployNodeConfigFieldsProviderUCloudUS3 from "./BizDeployNodeConfigFieldsProviderUCloudUS3";
import BizDeployNodeConfigFieldsProviderUniCloudWebHost from "./BizDeployNodeConfigFieldsProviderUniCloudWebHost";
import BizDeployNodeConfigFieldsProviderUpyunCDN from "./BizDeployNodeConfigFieldsProviderUpyunCDN";
import BizDeployNodeConfigFieldsProviderUpyunFile from "./BizDeployNodeConfigFieldsProviderUpyunFile";
import BizDeployNodeConfigFieldsProviderVolcEngineALB from "./BizDeployNodeConfigFieldsProviderVolcEngineALB";
import BizDeployNodeConfigFieldsProviderVolcEngineCDN from "./BizDeployNodeConfigFieldsProviderVolcEngineCDN";
import BizDeployNodeConfigFieldsProviderVolcEngineCertCenter from "./BizDeployNodeConfigFieldsProviderVolcEngineCertCenter";
import BizDeployNodeConfigFieldsProviderVolcEngineCLB from "./BizDeployNodeConfigFieldsProviderVolcEngineCLB";
import BizDeployNodeConfigFieldsProviderVolcEngineDCDN from "./BizDeployNodeConfigFieldsProviderVolcEngineDCDN";
import BizDeployNodeConfigFieldsProviderVolcEngineImageX from "./BizDeployNodeConfigFieldsProviderVolcEngineImageX";
import BizDeployNodeConfigFieldsProviderVolcEngineLive from "./BizDeployNodeConfigFieldsProviderVolcEngineLive";
import BizDeployNodeConfigFieldsProviderVolcEngineTOS from "./BizDeployNodeConfigFieldsProviderVolcEngineTOS";
import BizDeployNodeConfigFieldsProviderVolcEngineVOD from "./BizDeployNodeConfigFieldsProviderVolcEngineVOD.tsx";
import BizDeployNodeConfigFieldsProviderVolcEngineWAF from "./BizDeployNodeConfigFieldsProviderVolcEngineWAF.tsx";
import BizDeployNodeConfigFieldsProviderWangsuCDN from "./BizDeployNodeConfigFieldsProviderWangsuCDN";
import BizDeployNodeConfigFieldsProviderWangsuCDNPro from "./BizDeployNodeConfigFieldsProviderWangsuCDNPro";
import BizDeployNodeConfigFieldsProviderWangsuCertificate from "./BizDeployNodeConfigFieldsProviderWangsuCertificate";
import BizDeployNodeConfigFieldsProviderWebhook from "./BizDeployNodeConfigFieldsProviderWebhook";

const providerComponentMap: Partial<Record<DeploymentProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [DEPLOYMENT_PROVIDERS["1PANEL_CONSOLE"]]: BizDeployNodeConfigFieldsProvider1PanelConsole,
  [DEPLOYMENT_PROVIDERS["1PANEL"]]: BizDeployNodeConfigFieldsProvider1Panel,
  [DEPLOYMENT_PROVIDERS.ALIYUN_ALB]: BizDeployNodeConfigFieldsProviderAliyunALB,
  [DEPLOYMENT_PROVIDERS.ALIYUN_APIGW]: BizDeployNodeConfigFieldsProviderAliyunAPIGW,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CAS]: BizDeployNodeConfigFieldsProviderAliyunCAS,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CAS_DEPLOY]: BizDeployNodeConfigFieldsProviderAliyunCASDeploy,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CLB]: BizDeployNodeConfigFieldsProviderAliyunCLB,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CDN]: BizDeployNodeConfigFieldsProviderAliyunCDN,
  [DEPLOYMENT_PROVIDERS.ALIYUN_DCDN]: BizDeployNodeConfigFieldsProviderAliyunDCDN,
  [DEPLOYMENT_PROVIDERS.ALIYUN_DDOSPRO]: BizDeployNodeConfigFieldsProviderAliyunDDoSPro,
  [DEPLOYMENT_PROVIDERS.ALIYUN_ESA]: BizDeployNodeConfigFieldsProviderAliyunESA,
  [DEPLOYMENT_PROVIDERS.ALIYUN_FC]: BizDeployNodeConfigFieldsProviderAliyunFC,
  [DEPLOYMENT_PROVIDERS.ALIYUN_GA]: BizDeployNodeConfigFieldsProviderAliyunGA,
  [DEPLOYMENT_PROVIDERS.ALIYUN_LIVE]: BizDeployNodeConfigFieldsProviderAliyunLive,
  [DEPLOYMENT_PROVIDERS.ALIYUN_NLB]: BizDeployNodeConfigFieldsProviderAliyunNLB,
  [DEPLOYMENT_PROVIDERS.ALIYUN_OSS]: BizDeployNodeConfigFieldsProviderAliyunOSS,
  [DEPLOYMENT_PROVIDERS.ALIYUN_VOD]: BizDeployNodeConfigFieldsProviderAliyunVOD,
  [DEPLOYMENT_PROVIDERS.ALIYUN_WAF]: BizDeployNodeConfigFieldsProviderAliyunWAF,
  [DEPLOYMENT_PROVIDERS.APISIX]: BizDeployNodeConfigFieldsProviderAPISIX,
  [DEPLOYMENT_PROVIDERS.AWS_ACM]: BizDeployNodeConfigFieldsProviderAWSACM,
  [DEPLOYMENT_PROVIDERS.AWS_CLOUDFRONT]: BizDeployNodeConfigFieldsProviderAWSCloudFront,
  [DEPLOYMENT_PROVIDERS.AWS_IAM]: BizDeployNodeConfigFieldsProviderAWSIAM,
  [DEPLOYMENT_PROVIDERS.AZURE_KEYVAULT]: BizDeployNodeConfigFieldsProviderAzureKeyVault,
  [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_APPBLB]: BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB,
  [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_BLB]: BizDeployNodeConfigFieldsProviderBaiduCloudBLB,
  [DEPLOYMENT_PROVIDERS.BAIDUCLOUD_CDN]: BizDeployNodeConfigFieldsProviderBaiduCloudCDN,
  [DEPLOYMENT_PROVIDERS.BAISHAN_CDN]: BizDeployNodeConfigFieldsProviderBaishanCDN,
  [DEPLOYMENT_PROVIDERS.BAOTAPANEL_CONSOLE]: BizDeployNodeConfigFieldsProviderBaotaPanelConsole,
  [DEPLOYMENT_PROVIDERS.BAOTAPANEL]: BizDeployNodeConfigFieldsProviderBaotaPanel,
  [DEPLOYMENT_PROVIDERS.BAOTAPANELGO_CONSOLE]: BizDeployNodeConfigFieldsProviderBaotaPanelGoConsole,
  [DEPLOYMENT_PROVIDERS.BAOTAPANELGO]: BizDeployNodeConfigFieldsProviderBaotaPanelGo,
  [DEPLOYMENT_PROVIDERS.BAOTAWAF]: BizDeployNodeConfigFieldsProviderBaotaWAF,
  [DEPLOYMENT_PROVIDERS.BAOTAWAF_CONSOLE]: BizDeployNodeConfigFieldsProviderBaotaWAFConsole,
  [DEPLOYMENT_PROVIDERS.BUNNY_CDN]: BizDeployNodeConfigFieldsProviderBunnyCDN,
  [DEPLOYMENT_PROVIDERS.BYTEPLUS_CDN]: BizDeployNodeConfigFieldsProviderBytePlusCDN,
  [DEPLOYMENT_PROVIDERS.CDNFLY]: BizDeployNodeConfigFieldsProviderCdnfly,
  [DEPLOYMENT_PROVIDERS.CPANEL]: BizDeployNodeConfigFieldsProviderCPanel,
  [DEPLOYMENT_PROVIDERS.CTCCCLOUD_AO]: BizDeployNodeConfigFieldsProviderCTCCCloudAO,
  [DEPLOYMENT_PROVIDERS.CTCCCLOUD_CDN]: BizDeployNodeConfigFieldsProviderCTCCCloudCDN,
  [DEPLOYMENT_PROVIDERS.CTCCCLOUD_ELB]: BizDeployNodeConfigFieldsProviderCTCCCloudELB,
  [DEPLOYMENT_PROVIDERS.CTCCCLOUD_FAAS]: BizDeployNodeConfigFieldsProviderCTCCCloudFaaS,
  [DEPLOYMENT_PROVIDERS.CTCCCLOUD_ICDN]: BizDeployNodeConfigFieldsProviderCTCCCloudICDN,
  [DEPLOYMENT_PROVIDERS.CTCCCLOUD_LVDN]: BizDeployNodeConfigFieldsProviderCTCCCloudLVDN,
  [DEPLOYMENT_PROVIDERS.DOGECLOUD_CDN]: BizDeployNodeConfigFieldsProviderDogeCloudCDN,
  [DEPLOYMENT_PROVIDERS.FLEXCDN]: BizDeployNodeConfigFieldsProviderFlexCDN,
  [DEPLOYMENT_PROVIDERS.GCORE_CDN]: BizDeployNodeConfigFieldsProviderGcoreCDN,
  [DEPLOYMENT_PROVIDERS.GOEDGE]: BizDeployNodeConfigFieldsProviderGoEdge,
  [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_CDN]: BizDeployNodeConfigFieldsProviderHuaweiCloudCDN,
  [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_ELB]: BizDeployNodeConfigFieldsProviderHuaweiCloudELB,
  [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_OBS]: BizDeployNodeConfigFieldsProviderHuaweiCloudOBS,
  [DEPLOYMENT_PROVIDERS.HUAWEICLOUD_WAF]: BizDeployNodeConfigFieldsProviderHuaweiCloudWAF,
  [DEPLOYMENT_PROVIDERS.JDCLOUD_ALB]: BizDeployNodeConfigFieldsProviderJDCloudALB,
  [DEPLOYMENT_PROVIDERS.JDCLOUD_CDN]: BizDeployNodeConfigFieldsProviderJDCloudCDN,
  [DEPLOYMENT_PROVIDERS.JDCLOUD_LIVE]: BizDeployNodeConfigFieldsProviderJDCloudLive,
  [DEPLOYMENT_PROVIDERS.JDCLOUD_VOD]: BizDeployNodeConfigFieldsProviderJDCloudVOD,
  [DEPLOYMENT_PROVIDERS.KONG]: BizDeployNodeConfigFieldsProviderKong,
  [DEPLOYMENT_PROVIDERS.KUBERNETES_SECRET]: BizDeployNodeConfigFieldsProviderKubernetesSecret,
  [DEPLOYMENT_PROVIDERS.KSYUN_CDN]: BizDeployNodeConfigFieldsProviderKsyunCDN,
  [DEPLOYMENT_PROVIDERS.LECDN]: BizDeployNodeConfigFieldsProviderLeCDN,
  [DEPLOYMENT_PROVIDERS.LOCAL]: BizDeployNodeConfigFieldsProviderLocal,
  [DEPLOYMENT_PROVIDERS.MOHUA_MVH]: BizDeployNodeConfigFieldsProviderMohuaMVH,
  [DEPLOYMENT_PROVIDERS.NETLIFY]: BizDeployNodeConfigFieldsProviderNetlify,
  [DEPLOYMENT_PROVIDERS.NGINXPROXYMANAGER]: BizDeployNodeConfigFieldsProviderNginxProxyManager,
  [DEPLOYMENT_PROVIDERS.PROXMOXVE]: BizDeployNodeConfigFieldsProviderProxmoxVE,
  [DEPLOYMENT_PROVIDERS.QINIU_CDN]: BizDeployNodeConfigFieldsProviderQiniuCDN,
  [DEPLOYMENT_PROVIDERS.QINIU_KODO]: BizDeployNodeConfigFieldsProviderQiniuKodo,
  [DEPLOYMENT_PROVIDERS.QINIU_PILI]: BizDeployNodeConfigFieldsProviderQiniuPili,
  [DEPLOYMENT_PROVIDERS.RAINYUN_RCDN]: BizDeployNodeConfigFieldsProviderRainYunRCDN,
  [DEPLOYMENT_PROVIDERS.RATPANEL]: BizDeployNodeConfigFieldsProviderRatPanel,
  [DEPLOYMENT_PROVIDERS.S3]: BizDeployNodeConfigFieldsProviderS3,
  [DEPLOYMENT_PROVIDERS.SAFELINE]: BizDeployNodeConfigFieldsProviderSafeLine,
  [DEPLOYMENT_PROVIDERS.SSH]: BizDeployNodeConfigFieldsProviderSSH,
  [DEPLOYMENT_PROVIDERS.SYNOLOGYDSM]: BizDeployNodeConfigFieldsProviderSynologyDSM,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CDN]: BizDeployNodeConfigFieldsProviderTencentCloudCDN,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CLB]: BizDeployNodeConfigFieldsProviderTencentCloudCLB,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_COS]: BizDeployNodeConfigFieldsProviderTencentCloudCOS,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CSS]: BizDeployNodeConfigFieldsProviderTencentCloudCSS,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_ECDN]: BizDeployNodeConfigFieldsProviderTencentCloudECDN,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_EO]: BizDeployNodeConfigFieldsProviderTencentCloudEO,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_GAAP]: BizDeployNodeConfigFieldsProviderTencentCloudGAAP,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SCF]: BizDeployNodeConfigFieldsProviderTencentCloudSCF,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL]: BizDeployNodeConfigFieldsProviderTencentCloudSSL,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_DEPLOY]: BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_UPDATE]: BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_VOD]: BizDeployNodeConfigFieldsProviderTencentCloudVOD,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_WAF]: BizDeployNodeConfigFieldsProviderTencentCloudWAF,
  [DEPLOYMENT_PROVIDERS.UCLOUD_UALB]: BizDeployNodeConfigFieldsProviderUCloudUALB,
  [DEPLOYMENT_PROVIDERS.UCLOUD_UCDN]: BizDeployNodeConfigFieldsProviderUCloudUCDN,
  [DEPLOYMENT_PROVIDERS.UCLOUD_UCLB]: BizDeployNodeConfigFieldsProviderUCloudUCLB,
  [DEPLOYMENT_PROVIDERS.UCLOUD_UEWAF]: BizDeployNodeConfigFieldsProviderUCloudUEWAF,
  [DEPLOYMENT_PROVIDERS.UCLOUD_UPATHX]: BizDeployNodeConfigFieldsProviderUCloudUPathX,
  [DEPLOYMENT_PROVIDERS.UCLOUD_US3]: BizDeployNodeConfigFieldsProviderUCloudUS3,
  [DEPLOYMENT_PROVIDERS.UNICLOUD_WEBHOST]: BizDeployNodeConfigFieldsProviderUniCloudWebHost,
  [DEPLOYMENT_PROVIDERS.UPYUN_CDN]: BizDeployNodeConfigFieldsProviderUpyunCDN,
  [DEPLOYMENT_PROVIDERS.UPYUN_FILE]: BizDeployNodeConfigFieldsProviderUpyunFile,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_ALB]: BizDeployNodeConfigFieldsProviderVolcEngineALB,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_CDN]: BizDeployNodeConfigFieldsProviderVolcEngineCDN,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_CERTCENTER]: BizDeployNodeConfigFieldsProviderVolcEngineCertCenter,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_CLB]: BizDeployNodeConfigFieldsProviderVolcEngineCLB,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_DCDN]: BizDeployNodeConfigFieldsProviderVolcEngineDCDN,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_IMAGEX]: BizDeployNodeConfigFieldsProviderVolcEngineImageX,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_LIVE]: BizDeployNodeConfigFieldsProviderVolcEngineLive,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_TOS]: BizDeployNodeConfigFieldsProviderVolcEngineTOS,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_VOD]: BizDeployNodeConfigFieldsProviderVolcEngineVOD,
  [DEPLOYMENT_PROVIDERS.VOLCENGINE_WAF]: BizDeployNodeConfigFieldsProviderVolcEngineWAF,
  [DEPLOYMENT_PROVIDERS.WANGSU_CDN]: BizDeployNodeConfigFieldsProviderWangsuCDN,
  [DEPLOYMENT_PROVIDERS.WANGSU_CDNPRO]: BizDeployNodeConfigFieldsProviderWangsuCDNPro,
  [DEPLOYMENT_PROVIDERS.WANGSU_CERTIFICATE]: BizDeployNodeConfigFieldsProviderWangsuCertificate,
  [DEPLOYMENT_PROVIDERS.WEBHOOK]: BizDeployNodeConfigFieldsProviderWebhook,
};

const useComponent = (provider: string, { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }) => {
  const initComponent = () => {
    const Component = providerComponentMap[provider as DeploymentProviderType];
    if (!Component) return null;

    const props = initProps?.(provider);
    if (props) {
      return <Component {...props} />;
    }

    return <Component />;
  };

  const [component, setComponent] = useState(() => initComponent());

  useEffect(() => setComponent(initComponent()), [provider]);
  useEffect(() => setComponent(initComponent()), deps);

  return component;
};

const _default = {
  useComponent,
};

export default _default;
