import { useTranslation } from "react-i18next";
import { Avatar, Select, Typography, theme } from "antd";

import { type PurgeProvider } from "@/domain/provider";
import { getPurgeProvider, listPurgeProviders } from "@/domain/purgeProvider";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { matchSearchOption } from "@/utils/search";

import { type SharedSelectProps, useSelectDataSource } from "./_shared";

export interface PurgeProviderSelectProps extends SharedSelectProps<PurgeProvider> {
  showAvailability?: boolean;
}

const PurgeProviderSelect = ({ showAvailability = false, onFilter, ...props }: PurgeProviderSelectProps) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  usePluginCatalogStore((s) => s.version);

  const dataSource = useSelectDataSource({
    dataSource: listPurgeProviders(),
    filters: [onFilter!],
  });

  const options = (() => {
    const convert = (providers: PurgeProvider[]): Array<{ key: string; value: string; label: string; data: PurgeProvider }> => {
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
            options: convert(dataSource.available),
          },
          {
            label: t("provider.text.unavailable_group"),
            options: convert(dataSource.unavailable),
          },
        ].filter((group) => group.options.length > 0)
      : convert(dataSource.filtered);
  })();

  const renderOption = (key: string) => {
    const provider = getPurgeProvider(key);
    return (
      <div className="flex items-center gap-2 truncate overflow-hidden">
        {provider ? <Avatar shape="square" src={provider?.icon} size="small" /> : null}
        <Typography.Text ellipsis>{provider ? t(provider.name) : key}</Typography.Text>
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

export default PurgeProviderSelect;
