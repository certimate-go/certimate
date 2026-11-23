import { useTranslation } from "react-i18next";
import { useControllableValue, useMount } from "ahooks";
import { Avatar, Select, Typography, theme } from "antd";

import { type CAProvider, caProvidersMap } from "@/domain/provider";
import { useZustandShallowSelector } from "@/hooks";
import { useSSLProviderSettingsStore } from "@/stores/settings";
import { matchSearchOption } from "@/utils/search";

import { type SharedSelectProps, useSelectDataSource } from "./_shared";

export interface CAProviderSelectProps extends SharedSelectProps<CAProvider> {
  showAvailability?: boolean;
  showDefault?: boolean;
}

const CAProviderSelect = ({ showAvailability, showDefault, onFilter, ...props }: CAProviderSelectProps) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { settings: sslProviderSettings, loadSettings: loadSSLProviderSettings } = useSSLProviderSettingsStore(
    useZustandShallowSelector(["settings", "loadSettings"])
  );
  useMount(() => loadSSLProviderSettings(false));

  const [value, setValue] = useControllableValue<string | undefined>(props, {
    valuePropName: "value",
    defaultValuePropName: "defaultValue",
    trigger: "onChange",
  });

  const defaultCAProvider = caProvidersMap.get(sslProviderSettings.provider);

  const dataSource = useSelectDataSource({
    dataSource: Array.from(caProvidersMap.values()),
    filters: [onFilter!],
  });
  const options = (() => {
    const convert = (providers: CAProvider[]): Array<{ key: string; value: string; label: string; data: CAProvider }> => {
      return providers.map((provider) => ({
        key: provider.type,
        value: provider.type,
        label: t(provider.name),
        data: provider,
      }));
    };

    const defaultOption = {
      key: "",
      value: "",
      data: {} as CAProvider,
    };
    const plainOptions = convert(dataSource.filtered);
    const groupOptions = [
      {
        label: t("provider.text.available_group"),
        options: convert(dataSource.available),
      },
      {
        label: t("provider.text.unavailable_group"),
        options: convert(dataSource.unavailable),
      },
    ].filter((group) => group.options.length > 0);

    return showAvailability
      ? showDefault
        ? [{ label: t("provider.text.default_group"), options: [defaultOption] }, ...groupOptions]
        : groupOptions
      : showDefault
        ? [defaultOption, ...plainOptions]
        : plainOptions;
  })();

  const renderOption = (key: string) => {
    if (key === "") {
      return (
        <div className="flex items-center justify-between gap-4">
          <div className="flex-1 truncate">
            <Typography.Text ellipsis>{showAvailability ? t("provider.text.default_ca_in_group") : t("provider.text.default_ca")}</Typography.Text>
          </div>
          {defaultCAProvider && (
            <Typography.Text className="text-xs" type="secondary" ellipsis>
              {t(defaultCAProvider.name)}
            </Typography.Text>
          )}
        </div>
      );
    }

    const provider = caProvidersMap.get(key);
    return (
      <div className="flex items-center gap-2 truncate overflow-hidden">
        <Avatar shape="square" src={provider?.icon} size="small" />
        <Typography.Text ellipsis>{t(provider?.name ?? "")}</Typography.Text>
      </div>
    );
  };

  const handleChange = (value: string) => {
    setValue((_) => (value !== "" ? value : void 0));
  };

  return (
    <Select
      {...props}
      labelRender={({ value }) => {
        if (value != null && value !== "") {
          return renderOption(value as string);
        }

        return <span style={{ color: themeToken.colorTextPlaceholder }}>{props.placeholder}</span>;
      }}
      options={options}
      optionLabelProp={void 0}
      optionRender={(option) => renderOption(option.data.value as string)}
      showSearch={{
        filterOption: (inputValue, option) => {
          if (option?.value === "") return true; // 始终显示系统默认项

          return matchSearchOption(inputValue, option!);
        },
      }}
      value={value}
      onChange={handleChange}
      onSelect={handleChange}
    />
  );
};

export default CAProviderSelect;
