import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useControllableValue } from "ahooks";
import { Avatar, Select, Typography, theme } from "antd";

import { type CertsourceProvider, certsourceProvidersMap } from "@/domain/provider";
import { matchSearchOption } from "@/utils/search";

import { type SharedSelectProps, useSelectDataSource } from "./_shared";

export interface CertsourceProviderSelectProps extends SharedSelectProps<CertsourceProvider> {
  showAvailability?: boolean;
}

const CertsourceProviderSelect = ({ showAvailability, onFilter, ...props }: CertsourceProviderSelectProps) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const [value, setValue] = useControllableValue<string | undefined>(props, {
    valuePropName: "value",
    defaultValuePropName: "defaultValue",
    trigger: "onChange",
  });

  const dataSources = useSelectDataSource({
    dataSource: Array.from(certsourceProvidersMap.values()),
    filters: [onFilter!],
  });
  const options = useMemo(() => {
    const convert = (providers: CertsourceProvider[]): Array<{ key: string; value: string; label: string; data: CertsourceProvider }> => {
      return providers.map((provider) => ({
        key: provider.type,
        value: provider.type,
        label: t(provider.name),
        data: provider,
      }));
    };

    const plainOptions = convert(dataSources.filtered);
    const groupOptions = [
      {
        label: t("provider.text.available_group"),
        options: convert(dataSources.available),
      },
      {
        label: t("provider.text.unavailable_group"),
        options: convert(dataSources.unavailable),
      },
    ].filter((group) => group.options.length > 0);

    return showAvailability ? groupOptions : plainOptions;
  }, [showAvailability, dataSources]);

  const renderOption = (key: string) => {
    const provider = certsourceProvidersMap.get(key);
    return (
      <div className="flex items-center gap-2 truncate overflow-hidden">
        <Avatar shape="square" src={provider?.icon} size="small" />
        <Typography.Text ellipsis>{t(provider?.name ?? "")}</Typography.Text>
      </div>
    );
  };

  const handleChange = (value: string) => {
    setValue((_) => value);
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
        filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
      }}
      value={value}
      onChange={handleChange}
      onSelect={handleChange}
    />
  );
};

export default CertsourceProviderSelect;
