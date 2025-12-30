import { useEffect, useState } from "react";

import { type ACMEDns01ProviderType, type ACMEHttp01ProviderType, ACME_DNS01_PROVIDERS, ACME_HTTP01_PROVIDERS } from "@/domain/provider";

import BizApplyNodeConfigFieldsProviderAliyunESA from "./BizApplyNodeConfigFieldsProviderAliyunESA";
import BizApplyNodeConfigFieldsProviderAWSRoute53 from "./BizApplyNodeConfigFieldsProviderAWSRoute53";
import BizApplyNodeConfigFieldsProviderHuaweiCloudDNS from "./BizApplyNodeConfigFieldsProviderHuaweiCloudDNS";
import BizApplyNodeConfigFieldsProviderJDCloudDNS from "./BizApplyNodeConfigFieldsProviderJDCloudDNS";
import BizApplyNodeConfigFieldsProviderLocal from "./BizApplyNodeConfigFieldsProviderLocal";
import BizApplyNodeConfigFieldsProviderS3 from "./BizApplyNodeConfigFieldsProviderS3";
import BizApplyNodeConfigFieldsProviderSSH from "./BizApplyNodeConfigFieldsProviderSSH";

const acmeDns01ProviderComponentMap: Partial<Record<ACMEDns01ProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACME_DNS01_PROVIDERS.ALIYUN_ESA]: BizApplyNodeConfigFieldsProviderAliyunESA,
  [ACME_DNS01_PROVIDERS.AWS]: BizApplyNodeConfigFieldsProviderAWSRoute53,
  [ACME_DNS01_PROVIDERS.AWS_ROUTE53]: BizApplyNodeConfigFieldsProviderAWSRoute53,
  [ACME_DNS01_PROVIDERS.HUAWEICLOUD]: BizApplyNodeConfigFieldsProviderHuaweiCloudDNS,
  [ACME_DNS01_PROVIDERS.HUAWEICLOUD_DNS]: BizApplyNodeConfigFieldsProviderHuaweiCloudDNS,
  [ACME_DNS01_PROVIDERS.JDCLOUD]: BizApplyNodeConfigFieldsProviderJDCloudDNS,
  [ACME_DNS01_PROVIDERS.JDCLOUD_DNS]: BizApplyNodeConfigFieldsProviderJDCloudDNS,
};

const acmeHttp01ProviderComponentMap: Partial<Record<ACMEHttp01ProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACME_HTTP01_PROVIDERS.LOCAL]: BizApplyNodeConfigFieldsProviderLocal,
  [ACME_HTTP01_PROVIDERS.S3]: BizApplyNodeConfigFieldsProviderS3,
  [ACME_HTTP01_PROVIDERS.SSH]: BizApplyNodeConfigFieldsProviderSSH,
};

const useComponent = (
  challenge: "dns-01" | "http-01",
  provider: string,
  { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }
) => {
  const initComponent = () => {
    const Component =
      challenge === "dns-01"
        ? acmeDns01ProviderComponentMap[provider as ACMEDns01ProviderType]
        : challenge === "http-01"
          ? acmeHttp01ProviderComponentMap[provider as ACMEHttp01ProviderType]
          : void 0;
    if (!Component) return null;

    const props = initProps?.(provider);
    if (props) {
      return <Component {...props} />;
    }

    return <Component />;
  };

  const [component, setComponent] = useState(() => initComponent());

  useEffect(() => setComponent(initComponent()), [challenge, provider]);
  useEffect(() => setComponent(initComponent()), deps);

  return component;
};

const _default = {
  useComponent,
};

export default _default;
