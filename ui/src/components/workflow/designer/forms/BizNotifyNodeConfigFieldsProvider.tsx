import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import { NOTIFICATION_PROVIDERS, notificationProvidersMap } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import BizNotifyNodeConfigFieldsProviderEmail from "./BizNotifyNodeConfigFieldsProviderEmail";
import BizNotifyNodeConfigFieldsProviderWebhook from "./BizNotifyNodeConfigFieldsProviderWebhook";
import SchemaConfigFields from "./SchemaConfigFields";

const BizNotifyNodeConfigFieldsProvider = ({ provider }: { provider?: string }) => {
  const [envelope, setEnvelope] = useState<ProviderSchemaEnvelope | null>(null);
  const catalogVersion = usePluginCatalogStore((s) => s.version);
  const isPlugin = !!provider && notificationProvidersMap.get(provider)?.source === "plugin";

  useEffect(() => {
    if (!provider || !isPlugin) return;

    let cancelled = false;
    void getProviderSchema(provider, "notify").then((schema) => {
      if (!cancelled && schema) setEnvelope(schema);
    });
    return () => {
      cancelled = true;
    };
  }, [provider, isPlugin, catalogVersion]);

  if (provider === NOTIFICATION_PROVIDERS.EMAIL) return <BizNotifyNodeConfigFieldsProviderEmail />;
  if (provider === NOTIFICATION_PROVIDERS.WEBHOOK) return <BizNotifyNodeConfigFieldsProviderWebhook />;
  if (isPlugin && envelope && envelope.provider === provider) return <SchemaConfigFields envelope={envelope} />;

  return null;
};

export default BizNotifyNodeConfigFieldsProvider;
