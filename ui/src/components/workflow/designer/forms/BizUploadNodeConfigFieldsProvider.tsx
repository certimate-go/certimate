import { useEffect, useState } from "react";

import { CERTSOURCE_PROVIDERS, type CertsourceProviderType } from "@/domain/provider";

import BizUploadNodeConfigFieldsProviderAutoSSL from "./BizUploadNodeConfigFieldsProviderAutoSSL";

const providerComponentMap: Partial<Record<CertsourceProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [CERTSOURCE_PROVIDERS.AUTOSSL]: BizUploadNodeConfigFieldsProviderAutoSSL,
};

const useComponent = (provider: string, { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }) => {
  const initComponent = () => {
    const Component = providerComponentMap[provider as CertsourceProviderType];
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
