import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import { resolvePurgeProvider } from "@/domain/purgeProvider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import SchemaConfigFields from "./SchemaConfigFields";

const providerComponentMap: Partial<Record<string, React.ComponentType<any>>> = {};

const useComponent = (provider: string, { initProps, deps = [] }: { initProps?: (provider: string) => any; deps?: unknown[] }) => {
  const [envelope, setEnvelope] = useState<ProviderSchemaEnvelope | null>(null);
  const catalogVersion = usePluginCatalogStore((s) => s.version);

  useEffect(() => {
    let cancelled = false;
    setEnvelope(null);
    if (!provider) {
      return;
    }
    void getProviderSchema(resolvePurgeProvider(provider), "purge").then((env) => {
      if (!cancelled) {
        setEnvelope(env);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [provider, catalogVersion]);

  const initComponent = () => {
    const Component = providerComponentMap[provider];
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
