import { memo, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Select, type SelectProps, Typography, theme } from "antd";

import { type DeploymentProvider, deploymentProvidersMap } from "@/domain/provider";

export type DeploymentProviderSelectProps = Omit<
  SelectProps,
  "filterOption" | "filterSort" | "labelRender" | "options" | "optionFilterProp" | "optionLabelProp" | "optionRender"
> & {
  filter?: (record: DeploymentProvider) => boolean;
};

const DeploymentProviderSelect = ({ filter, ...props }: DeploymentProviderSelectProps) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const options = useMemo<Array<{ key: string; value: string; label: string; data: DeploymentProvider }>>(() => {
    return Array.from(deploymentProvidersMap.values())
      .filter((provider) => {
        if (filter) {
          return filter(provider);
        }

        return true;
      })
      .map((provider) => ({
        key: provider.type,
        value: provider.type,
        label: t(provider.name),
        data: provider,
      }));
  }, [filter]);

  const renderOption = (key: string) => {
    const provider = deploymentProvidersMap.get(key);
    return (
      <div className="flex items-center gap-2 overflow-hidden truncate">
        <Avatar shape="square" src={provider?.icon} size="small" />
        <Typography.Text ellipsis>{t(provider?.name ?? "")}</Typography.Text>
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
      optionFilterProp={undefined}
      optionLabelProp={undefined}
      optionRender={(option) => renderOption(option.data.value)}
    />
  );
};

export default memo(DeploymentProviderSelect);
