import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Select, Tag, Typography, theme } from "antd";

import Show from "@/components/Show";
import { ACCESS_USAGES, type AccessProvider, type AccessUsageType, accessProvidersMap, getAccessProvider } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { matchSearchOption } from "@/utils/search";

import { type SharedSelectProps } from "./_shared";

export interface AccessProviderSelectProps extends SharedSelectProps<AccessProvider> {
  showOptionTags?: boolean | { [key in AccessUsageType | "builtin"]?: boolean };
}

const AccessProviderSelect = ({ showOptionTags, onFilter, ...props }: AccessProviderSelectProps = { showOptionTags: true }) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const showOptionTagForDNS = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags?.[ACCESS_USAGES.DNS] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForHosting = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags?.[ACCESS_USAGES.HOSTING] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForCA = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags?.[ACCESS_USAGES.CA] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForNotification = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags?.[ACCESS_USAGES.NOTIFICATION] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForBuiltin = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags?.["builtin"] : !!showOptionTags;
  }, [showOptionTags]);

  usePluginCatalogStore((state) => state.version);
  const options = Array.from(accessProvidersMap.values())
    .filter((provider) => {
      if (onFilter) {
        return onFilter(provider.type, provider);
      }

      return true;
    })
    .map((provider) => ({
      key: provider.type,
      value: provider.type,
      label: t(provider.name),
      disabled: provider.builtin,
      data: provider,
    }));

  const renderOption = (key: string) => {
    const provider = getAccessProvider(key);
    return (
      <div className="flex max-w-full items-center justify-between gap-4 overflow-hidden">
        <div className="flex items-center gap-2 truncate overflow-hidden">
          <Avatar shape="square" src={provider?.icon} size="small" />
          <Typography.Text className="flex-1 truncate overflow-hidden" type={provider?.builtin ? "secondary" : void 0} ellipsis>
            {t(provider?.name ?? key)}
          </Typography.Text>
          {provider?.source === "plugin" && !accessProvidersMap.has(key) && (
            <Typography.Text type="secondary">{t("plugin.market.status.not_installed")}</Typography.Text>
          )}
        </div>
        <div className="flex origin-right scale-80 items-center justify-center gap-1 whitespace-nowrap">
          <Show when={showOptionTagForBuiltin && provider?.builtin}>
            <Tag color="default">{t("access.props.provider.builtin")}</Tag>
          </Show>
          <Show when={showOptionTagForDNS && provider?.usages.includes(ACCESS_USAGES.DNS)}>
            <Tag color="#d93f0b">{t("access.props.provider.usage.dns")}</Tag>
          </Show>
          <Show when={showOptionTagForHosting && provider?.usages.includes(ACCESS_USAGES.HOSTING)}>
            <Tag color="#0052cc">{t("access.props.provider.usage.hosting")}</Tag>
          </Show>
          <Show when={showOptionTagForCA && provider?.usages.includes(ACCESS_USAGES.CA)}>
            <Tag color="#0e8a16">{t("access.props.provider.usage.ca")}</Tag>
          </Show>
          <Show when={showOptionTagForNotification && provider?.usages.includes(ACCESS_USAGES.NOTIFICATION)}>
            <Tag color="#1d76db">{t("access.props.provider.usage.notification")}</Tag>
          </Show>
        </div>
      </div>
    );
  };

  return (
    <Select
      {...props}
      labelRender={({ value }) => {
        if (typeof value === "string") {
          return renderOption(value);
        }

        return <span style={{ color: themeToken.colorTextPlaceholder }}>{props.placeholder}</span>;
      }}
      options={options}
      optionLabelProp={void 0}
      optionRender={(option) => {
        const provider = accessProvidersMap.get(option.data.value as string);
        const hasDeployers = provider?.deployers && provider.deployers.length > 0;
        return (
          <div className="flex flex-col gap-0.5">
            {renderOption(option.data.value)}
            {hasDeployers && (
              <Typography.Text className="pl-7" type="secondary" style={{ fontSize: 11 }}>
                {t("provider.access.usedBy")} {provider!.deployers!.slice(0, 3).join(", ")}
                {provider!.deployers!.length > 3 ? ` +${provider!.deployers!.length - 3} more` : ""}
              </Typography.Text>
            )}
          </div>
        );
      }}
      showSearch={{
        filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
        optionFilterProp: "label",
      }}
    />
  );
};

export default AccessProviderSelect;
