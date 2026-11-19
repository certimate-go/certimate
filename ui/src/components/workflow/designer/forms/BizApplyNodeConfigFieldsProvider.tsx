import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import { type ACMEDns01ProviderType, type ACMEHttp01ProviderType, ACME_DNS01_PROVIDERS, ACME_HTTP01_PROVIDERS, getACMEDns01Provider } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import BizApplyNodeConfigFieldsProviderAliyunESA from "./BizApplyNodeConfigFieldsProviderAliyunESA";
import BizApplyNodeConfigFieldsProviderAWSLightsail from "./BizApplyNodeConfigFieldsProviderAWSLightsail";
import BizApplyNodeConfigFieldsProviderAWSRoute53 from "./BizApplyNodeConfigFieldsProviderAWSRoute53";
import BizApplyNodeConfigFieldsProviderFTP from "./BizApplyNodeConfigFieldsProviderFTP";
import BizApplyNodeConfigFieldsProviderLocal from "./BizApplyNodeConfigFieldsProviderLocal";
import BizApplyNodeConfigFieldsProviderS3 from "./BizApplyNodeConfigFieldsProviderS3";
import BizApplyNodeConfigFieldsProviderSSH from "./BizApplyNodeConfigFieldsProviderSSH";
import SchemaConfigFields from "./SchemaConfigFields";

const acmeDns01ProviderComponentMap: Partial<Record<ACMEDns01ProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACME_DNS01_PROVIDERS.ALIYUN_ESA]: BizApplyNodeConfigFieldsProviderAliyunESA,
  [ACME_DNS01_PROVIDERS.AWS]: BizApplyNodeConfigFieldsProviderAWSRoute53,
  [ACME_DNS01_PROVIDERS.AWS_LIGHTSAIL]: BizApplyNodeConfigFieldsProviderAWSLightsail,
  [ACME_DNS01_PROVIDERS.AWS_ROUTE53]: BizApplyNodeConfigFieldsProviderAWSRoute53,
};

const acmeHttp01ProviderComponentMap: Partial<Record<ACMEHttp01ProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [ACME_HTTP01_PROVIDERS.LOCAL]: BizApplyNodeConfigFieldsProviderLocal,
  [ACME_HTTP01_PROVIDERS.FTP]: BizApplyNodeConfigFieldsProviderFTP,
  [ACME_HTTP01_PROVIDERS.S3]: BizApplyNodeConfigFieldsProviderS3,
  [ACME_HTTP01_PROVIDERS.SSH]: BizApplyNodeConfigFieldsProviderSSH,
};

const useComponent = (
  challenge: "dns-01" | "http-01" | "dns-persist-01",
  selectedProvider: string | null | undefined,
  { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }
) => {
  const [envelope, setEnvelope] = useState<ProviderSchemaEnvelope | null>(null);
  const catalogVersion = usePluginCatalogStore((s) => s.version);
  const provider = challenge === "dns-01" && selectedProvider ? (getACMEDns01Provider(selectedProvider)?.type ?? selectedProvider) : selectedProvider;

  useEffect(() => {
    let cancelled = false;
    setEnvelope(null);
    if (!provider || challenge !== "dns-01") {
      return;
    }
    void getProviderSchema(provider, "apply").then((env) => {
      if (!cancelled) {
        setEnvelope(env);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [challenge, provider, catalogVersion]);

  const initComponent = () => {
    if (!provider) return null;
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

  useEffect(() => {
    setComponent(envelope ? <SchemaConfigFields envelope={envelope} /> : initComponent());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [challenge, provider, envelope]);
  useEffect(() => {
    setComponent(envelope ? <SchemaConfigFields envelope={envelope} /> : initComponent());
  }, deps);

  return component;
};

const _default = {
  useComponent,
};

export default _default;
