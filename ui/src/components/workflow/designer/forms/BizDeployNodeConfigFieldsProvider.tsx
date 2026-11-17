import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import { DEPLOYMENT_PROVIDERS, type DeploymentProviderType } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import BizDeployNodeConfigFieldsProviderAliyunALB from "./BizDeployNodeConfigFieldsProviderAliyunALB";
import BizDeployNodeConfigFieldsProviderAliyunAPIGW from "./BizDeployNodeConfigFieldsProviderAliyunAPIGW";
import BizDeployNodeConfigFieldsProviderAliyunCAS from "./BizDeployNodeConfigFieldsProviderAliyunCAS";
import BizDeployNodeConfigFieldsProviderAliyunCASDeploy from "./BizDeployNodeConfigFieldsProviderAliyunCASDeploy";
import BizDeployNodeConfigFieldsProviderAliyunCLB from "./BizDeployNodeConfigFieldsProviderAliyunCLB";
import BizDeployNodeConfigFieldsProviderAliyunDCDN from "./BizDeployNodeConfigFieldsProviderAliyunDCDN";
import BizDeployNodeConfigFieldsProviderAliyunDDoSPro from "./BizDeployNodeConfigFieldsProviderAliyunDDoSPro";
import BizDeployNodeConfigFieldsProviderAliyunESA from "./BizDeployNodeConfigFieldsProviderAliyunESA";
import BizDeployNodeConfigFieldsProviderAliyunESASaaS from "./BizDeployNodeConfigFieldsProviderAliyunESASaaS";
import BizDeployNodeConfigFieldsProviderAliyunFC from "./BizDeployNodeConfigFieldsProviderAliyunFC";
import BizDeployNodeConfigFieldsProviderAliyunGA from "./BizDeployNodeConfigFieldsProviderAliyunGA";
import BizDeployNodeConfigFieldsProviderAliyunLive from "./BizDeployNodeConfigFieldsProviderAliyunLive";
import BizDeployNodeConfigFieldsProviderAliyunMSE from "./BizDeployNodeConfigFieldsProviderAliyunMSE";
import BizDeployNodeConfigFieldsProviderAliyunNLB from "./BizDeployNodeConfigFieldsProviderAliyunNLB";
import BizDeployNodeConfigFieldsProviderAliyunOSS from "./BizDeployNodeConfigFieldsProviderAliyunOSS";
import BizDeployNodeConfigFieldsProviderAliyunVOD from "./BizDeployNodeConfigFieldsProviderAliyunVOD";
import BizDeployNodeConfigFieldsProviderAliyunWAF from "./BizDeployNodeConfigFieldsProviderAliyunWAF";
import BizDeployNodeConfigFieldsProviderAWSACM from "./BizDeployNodeConfigFieldsProviderAWSACM";
import BizDeployNodeConfigFieldsProviderAWSALB from "./BizDeployNodeConfigFieldsProviderAWSALB";
import BizDeployNodeConfigFieldsProviderAWSAmplify from "./BizDeployNodeConfigFieldsProviderAWSAmplify";
import BizDeployNodeConfigFieldsProviderAWSAPIGateway from "./BizDeployNodeConfigFieldsProviderAWSAPIGateway";
import BizDeployNodeConfigFieldsProviderAWSCLB from "./BizDeployNodeConfigFieldsProviderAWSCLB";
import BizDeployNodeConfigFieldsProviderAWSCloudFront from "./BizDeployNodeConfigFieldsProviderAWSCloudFront";
import BizDeployNodeConfigFieldsProviderAWSIAM from "./BizDeployNodeConfigFieldsProviderAWSIAM";
import BizDeployNodeConfigFieldsProviderAWSNLB from "./BizDeployNodeConfigFieldsProviderAWSNLB";
import BizDeployNodeConfigFieldsProviderAzureKeyVault from "./BizDeployNodeConfigFieldsProviderAzureKeyVault";
import BizDeployNodeConfigFieldsProviderFTP from "./BizDeployNodeConfigFieldsProviderFTP";
import BizDeployNodeConfigFieldsProviderLocal from "./BizDeployNodeConfigFieldsProviderLocal";
import BizDeployNodeConfigFieldsProviderS3 from "./BizDeployNodeConfigFieldsProviderS3";
import BizDeployNodeConfigFieldsProviderSSH from "./BizDeployNodeConfigFieldsProviderSSH";
import BizDeployNodeConfigFieldsProviderTencentCloudCDN from "./BizDeployNodeConfigFieldsProviderTencentCloudCDN";
import BizDeployNodeConfigFieldsProviderTencentCloudCLB from "./BizDeployNodeConfigFieldsProviderTencentCloudCLB";
import BizDeployNodeConfigFieldsProviderTencentCloudCOS from "./BizDeployNodeConfigFieldsProviderTencentCloudCOS";
import BizDeployNodeConfigFieldsProviderTencentCloudCSS from "./BizDeployNodeConfigFieldsProviderTencentCloudCSS";
import BizDeployNodeConfigFieldsProviderTencentCloudDDoSPro from "./BizDeployNodeConfigFieldsProviderTencentCloudDDoSPro";
import BizDeployNodeConfigFieldsProviderTencentCloudECDN from "./BizDeployNodeConfigFieldsProviderTencentCloudECDN";
import BizDeployNodeConfigFieldsProviderTencentCloudEO from "./BizDeployNodeConfigFieldsProviderTencentCloudEO";
import BizDeployNodeConfigFieldsProviderTencentCloudEOMakers from "./BizDeployNodeConfigFieldsProviderTencentCloudEOMakers";
import BizDeployNodeConfigFieldsProviderTencentCloudGA2 from "./BizDeployNodeConfigFieldsProviderTencentCloudGA2";
import BizDeployNodeConfigFieldsProviderTencentCloudGAAP from "./BizDeployNodeConfigFieldsProviderTencentCloudGAAP";
import BizDeployNodeConfigFieldsProviderTencentCloudSCF from "./BizDeployNodeConfigFieldsProviderTencentCloudSCF";
import BizDeployNodeConfigFieldsProviderTencentCloudSSL from "./BizDeployNodeConfigFieldsProviderTencentCloudSSL";
import BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy from "./BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy";
import BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate from "./BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate";
import BizDeployNodeConfigFieldsProviderTencentCloudTSE from "./BizDeployNodeConfigFieldsProviderTencentCloudTSE";
import BizDeployNodeConfigFieldsProviderTencentCloudVOD from "./BizDeployNodeConfigFieldsProviderTencentCloudVOD";
import BizDeployNodeConfigFieldsProviderTencentCloudWAF from "./BizDeployNodeConfigFieldsProviderTencentCloudWAF";
import BizDeployNodeConfigFieldsProviderWebhook from "./BizDeployNodeConfigFieldsProviderWebhook";
import SchemaConfigFields from "./SchemaConfigFields";

const providerComponentMap: Partial<Record<DeploymentProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [DEPLOYMENT_PROVIDERS.ALIYUN_ALB]: BizDeployNodeConfigFieldsProviderAliyunALB,
  [DEPLOYMENT_PROVIDERS.ALIYUN_APIGW]: BizDeployNodeConfigFieldsProviderAliyunAPIGW,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CAS]: BizDeployNodeConfigFieldsProviderAliyunCAS,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CAS_DEPLOY]: BizDeployNodeConfigFieldsProviderAliyunCASDeploy,
  [DEPLOYMENT_PROVIDERS.ALIYUN_CLB]: BizDeployNodeConfigFieldsProviderAliyunCLB,
  [DEPLOYMENT_PROVIDERS.ALIYUN_DCDN]: BizDeployNodeConfigFieldsProviderAliyunDCDN,
  [DEPLOYMENT_PROVIDERS.ALIYUN_DDOSPRO]: BizDeployNodeConfigFieldsProviderAliyunDDoSPro,
  [DEPLOYMENT_PROVIDERS.ALIYUN_ESA]: BizDeployNodeConfigFieldsProviderAliyunESA,
  [DEPLOYMENT_PROVIDERS.ALIYUN_ESA_SAAS]: BizDeployNodeConfigFieldsProviderAliyunESASaaS,
  [DEPLOYMENT_PROVIDERS.ALIYUN_FC]: BizDeployNodeConfigFieldsProviderAliyunFC,
  [DEPLOYMENT_PROVIDERS.ALIYUN_GA]: BizDeployNodeConfigFieldsProviderAliyunGA,
  [DEPLOYMENT_PROVIDERS.ALIYUN_LIVE]: BizDeployNodeConfigFieldsProviderAliyunLive,
  [DEPLOYMENT_PROVIDERS.ALIYUN_MSE]: BizDeployNodeConfigFieldsProviderAliyunMSE,
  [DEPLOYMENT_PROVIDERS.ALIYUN_NLB]: BizDeployNodeConfigFieldsProviderAliyunNLB,
  [DEPLOYMENT_PROVIDERS.ALIYUN_OSS]: BizDeployNodeConfigFieldsProviderAliyunOSS,
  [DEPLOYMENT_PROVIDERS.ALIYUN_VOD]: BizDeployNodeConfigFieldsProviderAliyunVOD,
  [DEPLOYMENT_PROVIDERS.ALIYUN_WAF]: BizDeployNodeConfigFieldsProviderAliyunWAF,
  [DEPLOYMENT_PROVIDERS.AWS_ACM]: BizDeployNodeConfigFieldsProviderAWSACM,
  [DEPLOYMENT_PROVIDERS.AWS_ALB]: BizDeployNodeConfigFieldsProviderAWSALB,
  [DEPLOYMENT_PROVIDERS.AWS_AMPLIFY]: BizDeployNodeConfigFieldsProviderAWSAmplify,
  [DEPLOYMENT_PROVIDERS.AWS_APIGATEWAY]: BizDeployNodeConfigFieldsProviderAWSAPIGateway,
  [DEPLOYMENT_PROVIDERS.AWS_CLB]: BizDeployNodeConfigFieldsProviderAWSCLB,
  [DEPLOYMENT_PROVIDERS.AWS_CLOUDFRONT]: BizDeployNodeConfigFieldsProviderAWSCloudFront,
  [DEPLOYMENT_PROVIDERS.AWS_IAM]: BizDeployNodeConfigFieldsProviderAWSIAM,
  [DEPLOYMENT_PROVIDERS.AWS_NLB]: BizDeployNodeConfigFieldsProviderAWSNLB,
  [DEPLOYMENT_PROVIDERS.AZURE_KEYVAULT]: BizDeployNodeConfigFieldsProviderAzureKeyVault,
  [DEPLOYMENT_PROVIDERS.FTP]: BizDeployNodeConfigFieldsProviderFTP,
  [DEPLOYMENT_PROVIDERS.LOCAL]: BizDeployNodeConfigFieldsProviderLocal,
  [DEPLOYMENT_PROVIDERS.S3]: BizDeployNodeConfigFieldsProviderS3,
  [DEPLOYMENT_PROVIDERS.SSH]: BizDeployNodeConfigFieldsProviderSSH,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CDN]: BizDeployNodeConfigFieldsProviderTencentCloudCDN,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CLB]: BizDeployNodeConfigFieldsProviderTencentCloudCLB,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_COS]: BizDeployNodeConfigFieldsProviderTencentCloudCOS,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CSS]: BizDeployNodeConfigFieldsProviderTencentCloudCSS,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_DDOSPRO]: BizDeployNodeConfigFieldsProviderTencentCloudDDoSPro,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_ECDN]: BizDeployNodeConfigFieldsProviderTencentCloudECDN,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_EO]: BizDeployNodeConfigFieldsProviderTencentCloudEO,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_EOMAKERS]: BizDeployNodeConfigFieldsProviderTencentCloudEOMakers,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_GAAP]: BizDeployNodeConfigFieldsProviderTencentCloudGAAP,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_GA2]: BizDeployNodeConfigFieldsProviderTencentCloudGA2,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SCF]: BizDeployNodeConfigFieldsProviderTencentCloudSCF,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL]: BizDeployNodeConfigFieldsProviderTencentCloudSSL,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_DEPLOY]: BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_UPDATE]: BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_TSE]: BizDeployNodeConfigFieldsProviderTencentCloudTSE,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_VOD]: BizDeployNodeConfigFieldsProviderTencentCloudVOD,
  [DEPLOYMENT_PROVIDERS.TENCENTCLOUD_WAF]: BizDeployNodeConfigFieldsProviderTencentCloudWAF,
  [DEPLOYMENT_PROVIDERS.WEBHOOK]: BizDeployNodeConfigFieldsProviderWebhook,
};

const useComponent = (provider: string, { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }) => {
  const [envelope, setEnvelope] = useState<ProviderSchemaEnvelope | null>(null);
  const catalogVersion = usePluginCatalogStore((s) => s.version);

  useEffect(() => {
    let cancelled = false;
    setEnvelope(null);
    if (!provider) {
      return;
    }
    void getProviderSchema(provider, "deploy").then((env) => {
      if (!cancelled) {
        setEnvelope(env);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [provider, catalogVersion]);

  const initComponent = () => {
    const Component = providerComponentMap[provider as DeploymentProviderType];
    if (!Component) return null;

    const props = initProps?.(provider);
    if (props) {
      return <Component {...props} />;
    }

    return <Component />;
  };

  const [component, setComponent] = useState<React.ReactNode>(() => initComponent());

  useEffect(() => {
    setComponent(envelope ? <SchemaConfigFields envelope={envelope} /> : initComponent());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [provider, envelope]);
  useEffect(() => {
    setComponent(envelope ? <SchemaConfigFields envelope={envelope} /> : initComponent());
  }, deps);

  return component;
};

const _default = {
  useComponent,
};

export default _default;
