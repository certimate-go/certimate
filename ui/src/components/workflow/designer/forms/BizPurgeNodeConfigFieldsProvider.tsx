import { useEffect, useState } from "react";

import { PURGE_PROVIDERS, type PurgeProviderType } from "@/domain/provider";

import BizPurgeNodeConfigFieldsProviderAliyunCAS from "./BizPurgeNodeConfigFieldsProviderAliyunCAS";
import BizPurgeNodeConfigFieldsProviderAliyunCLB from "./BizPurgeNodeConfigFieldsProviderAliyunCLB";
import BizPurgeNodeConfigFieldsProviderAliyunESA from "./BizPurgeNodeConfigFieldsProviderAliyunESA";
import BizPurgeNodeConfigFieldsProviderHuaweiCloudELB from "./BizPurgeNodeConfigFieldsProviderHuaweiCloudELB";
import BizPurgeNodeConfigFieldsProviderHuaweiCloudWAF from "./BizPurgeNodeConfigFieldsProviderHuaweiCloudWAF";
import BizPurgeNodeConfigFieldsProviderTencentCloudGAAP from "./BizPurgeNodeConfigFieldsProviderTencentCloudGAAP";
import BizPurgeNodeConfigFieldsProviderTencentCloudSSL from "./BizPurgeNodeConfigFieldsProviderTencentCloudSSL";
import BizPurgeNodeConfigFieldsProviderTencentCloudTSE from "./BizPurgeNodeConfigFieldsProviderTencentCloudTSE";
import BizPurgeNodeConfigFieldsProviderUCloudUSSL from "./BizPurgeNodeConfigFieldsProviderUCloudUSSL";

const providerComponentMap: Partial<Record<PurgeProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [PURGE_PROVIDERS.ALIYUN_CAS]: BizPurgeNodeConfigFieldsProviderAliyunCAS,
  [PURGE_PROVIDERS.ALIYUN_CLB]: BizPurgeNodeConfigFieldsProviderAliyunCLB,
  [PURGE_PROVIDERS.ALIYUN_ESA]: BizPurgeNodeConfigFieldsProviderAliyunESA,
  [PURGE_PROVIDERS.HUAWEICLOUD_ELB]: BizPurgeNodeConfigFieldsProviderHuaweiCloudELB,
  [PURGE_PROVIDERS.HUAWEICLOUD_WAF]: BizPurgeNodeConfigFieldsProviderHuaweiCloudWAF,
  [PURGE_PROVIDERS.TENCENTCLOUD_GAAP]: BizPurgeNodeConfigFieldsProviderTencentCloudGAAP,
  [PURGE_PROVIDERS.TENCENTCLOUD_SSL]: BizPurgeNodeConfigFieldsProviderTencentCloudSSL,
  [PURGE_PROVIDERS.TENCENTCLOUD_TSE]: BizPurgeNodeConfigFieldsProviderTencentCloudTSE,
  [PURGE_PROVIDERS.UCLOUD_USSL]: BizPurgeNodeConfigFieldsProviderUCloudUSSL,
};

const useComponent = (provider: string, { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }) => {
  const initComponent = () => {
    const Component = providerComponentMap[provider as PurgeProviderType];
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
