import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Select, Typography, theme } from "antd";

import { type ACMEDns01Provider, ACME_DNS01_PROVIDERS, acmeDns01ProvidersMap, getACMEDns01Provider } from "@/domain/provider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { matchSearchOption } from "@/utils/search";

import { type SharedSelectProps, useSelectDataSource } from "./_shared";

export interface ACMEDns01ProviderSelectProps extends SharedSelectProps<ACMEDns01Provider> {
  showAvailability?: boolean;
}

const ACMEDns01ProviderSelect = ({ showAvailability, onFilter, ...props }: ACMEDns01ProviderSelectProps) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  usePluginCatalogStore((state) => state.version);
  const dataSources = useSelectDataSource({
    dataSource: Array.from(acmeDns01ProvidersMap.values()),
    filters: [onFilter!],
  });
  const options = useMemo(() => {
    const convert = (providers: ACMEDns01Provider[]): Array<{ key: string; value: string; label: string; data: ACMEDns01Provider }> => {
      return providers.map((provider) => ({
        key: provider.type,
        value: provider.type,
        label: t(provider.name),
        data: provider,
      }));
    };

    return showAvailability
      ? [
          {
            label: t("provider.text.available_group"),
            options: convert(dataSources.available),
          },
          {
            label: t("provider.text.unavailable_group"),
            options: convert(dataSources.unavailable),
          },
        ].filter((group) => group.options.length > 0)
      : convert(dataSources.filtered);
  }, [showAvailability, dataSources, t]);

  const renderOption = (key: string) => {
    const provider = getACMEDns01Provider(key);
    const needsPlugin = provider?.type === ACME_DNS01_PROVIDERS.CMCCCLOUD_DNS && !acmeDns01ProvidersMap.has(provider.type);
    return (
      <div className="flex items-center gap-2 truncate overflow-hidden">
        <Avatar shape="square" src={provider?.icon} size="small" />
        <Typography.Text ellipsis>{t(provider?.name ?? key)}</Typography.Text>
        {needsPlugin && <Typography.Text type="secondary">{t("plugin.market.status.not_installed")}</Typography.Text>}
      </div>
    );
  };

  return (
    <Select
      {...props}
      labelRender={({ value }) => {
        if (value != null) {
          return renderOption(value as string);
        }

        return <span style={{ color: themeToken.colorTextPlaceholder }}>{props.placeholder}</span>;
      }}
      options={options}
      optionLabelProp={void 0}
      optionRender={(option) => renderOption(option.data.value as string)}
      showSearch={{
        filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
      }}
    />
  );
};

export default ACMEDns01ProviderSelect;
