import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Select, Tag, Typography, theme } from "antd";

import Show from "@/components/Show";
import { ACCESS_PROVIDERS, ACCESS_USAGES, type AccessProvider, type AccessUsageType, accessProvidersMap } from "@/domain/provider";

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

  const options = useMemo<Array<{ key: string; value: string; label: string; data: AccessProvider }>>(() => {
    return Array.from(accessProvidersMap.values())
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
        disabled: provider.builtin || provider.disabled,
        data: provider,
      }));
  }, [onFilter]);

  const renderOption = (key: string) => {
    const provider = accessProvidersMap.get(key) ?? ({ type: "", name: "", icon: "", usages: [] } as unknown as AccessProvider);
    const showUnsupportedBadge = provider.type === ACCESS_PROVIDERS.DOCKERHOST && provider.disabled;
    return (
      <div className="flex max-w-full items-center justify-between gap-4 overflow-hidden">
        <div className="flex items-center gap-2 truncate overflow-hidden">
          <Avatar shape="square" src={provider.icon} size="small" />
          <Typography.Text
            className="flex-1 truncate overflow-hidden"
            type={provider.builtin || provider.disabled ? "secondary" : void 0}
            ellipsis
          >
            {t(provider.name)}
          </Typography.Text>
        </div>
        <div className="origin-right scale-75 whitespace-nowrap">
          <Show when={showUnsupportedBadge}>
            <Tag color="default">{t("provider.badge.unsupported")}</Tag>
          </Show>
          <Show when={showOptionTagForBuiltin && provider.builtin}>
            <Tag>{t("access.props.provider.builtin")}</Tag>
          </Show>
          <Show when={showOptionTagForDNS && provider.usages.includes(ACCESS_USAGES.DNS)}>
            <Tag color="orange">{t("access.props.provider.usage.dns")}</Tag>
          </Show>
          <Show when={showOptionTagForHosting && provider.usages.includes(ACCESS_USAGES.HOSTING)}>
            <Tag color="geekblue">{t("access.props.provider.usage.hosting")}</Tag>
          </Show>
          <Show when={showOptionTagForCA && provider.usages.includes(ACCESS_USAGES.CA)}>
            <Tag color="magenta">{t("access.props.provider.usage.ca")}</Tag>
          </Show>
          <Show when={showOptionTagForNotification && provider.usages.includes(ACCESS_USAGES.NOTIFICATION)}>
            <Tag color="cyan">{t("access.props.provider.usage.notification")}</Tag>
          </Show>
        </div>
      </div>
    );
  };

  return (
    <Select
      {...props}
      filterOption={(inputValue, option) => {
        if (!option) return false;

        const value = inputValue.toLowerCase();
        return option.value.toLowerCase().includes(value) || option.label.toLowerCase().includes(value);
      }}
      labelRender={({ value }) => {
        if (value != null) {
          return renderOption(value as string);
        }

        return <span style={{ color: themeToken.colorTextPlaceholder }}>{props.placeholder}</span>;
      }}
      options={options}
      optionFilterProp={void 0}
      optionLabelProp={void 0}
      optionRender={(option) => renderOption(option.data.value)}
    />
  );
};

export default AccessProviderSelect;
