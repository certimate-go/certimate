import { useEffect, useState } from "react";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import { NOTIFICATION_PROVIDERS, type NotificationProviderType } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import BizNotifyNodeConfigFieldsProviderDiscordBot from "./BizNotifyNodeConfigFieldsProviderDiscordBot";
import BizNotifyNodeConfigFieldsProviderEmail from "./BizNotifyNodeConfigFieldsProviderEmail";
import BizNotifyNodeConfigFieldsProviderMatrix from "./BizNotifyNodeConfigFieldsProviderMatrix";
import BizNotifyNodeConfigFieldsProviderMattermost from "./BizNotifyNodeConfigFieldsProviderMattermost";
import BizNotifyNodeConfigFieldsProviderSlackBot from "./BizNotifyNodeConfigFieldsProviderSlackBot";
import BizNotifyNodeConfigFieldsProviderTelegramBot from "./BizNotifyNodeConfigFieldsProviderTelegramBot";
import BizNotifyNodeConfigFieldsProviderWebhook from "./BizNotifyNodeConfigFieldsProviderWebhook";
import SchemaConfigFields from "./SchemaConfigFields";

const providerComponentMap: Partial<Record<NotificationProviderType, React.ComponentType<any>>> = {
  /*
    注意：如果追加新的子组件，请保持以 ASCII 排序。
    NOTICE: If you add new child component, please keep ASCII order.
    */
  [NOTIFICATION_PROVIDERS.DISCORDBOT]: BizNotifyNodeConfigFieldsProviderDiscordBot,
  [NOTIFICATION_PROVIDERS.EMAIL]: BizNotifyNodeConfigFieldsProviderEmail,
  [NOTIFICATION_PROVIDERS.MATRIX]: BizNotifyNodeConfigFieldsProviderMatrix,
  [NOTIFICATION_PROVIDERS.MATTERMOST]: BizNotifyNodeConfigFieldsProviderMattermost,
  [NOTIFICATION_PROVIDERS.SLACKBOT]: BizNotifyNodeConfigFieldsProviderSlackBot,
  [NOTIFICATION_PROVIDERS.TELEGRAMBOT]: BizNotifyNodeConfigFieldsProviderTelegramBot,
  [NOTIFICATION_PROVIDERS.WEBHOOK]: BizNotifyNodeConfigFieldsProviderWebhook,
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
    void getProviderSchema(provider, "notify").then((env) => {
      if (!cancelled) {
        setEnvelope(env);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [provider, catalogVersion]);

  const initComponent = () => {
    const Component = providerComponentMap[provider as NotificationProviderType];
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
