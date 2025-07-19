import { useEffect, useState } from "react";
import { Avatar, Select, type SelectProps, Typography, theme } from "antd";

import { type AccessModel } from "@/domain/access";
import { accessProvidersMap } from "@/domain/provider";
import { useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";

export type AccessTypeSelectProps = Omit<
  SelectProps,
  "filterOption" | "filterSort" | "labelRender" | "loading" | "options" | "optionFilterProp" | "optionLabelProp" | "optionRender"
> & {
  filter?: (record: AccessModel) => boolean;
};

const AccessSelect = ({ filter, ...props }: AccessTypeSelectProps) => {
  const { token: themeToken } = theme.useToken();

  const { accesses, loadedAtOnce, fetchAccesses } = useAccessesStore(useZustandShallowSelector(["accesses", "loadedAtOnce", "fetchAccesses"]));
  useEffect(() => {
    fetchAccesses();
  }, []);

  const [options, setOptions] = useState<Array<{ key: string; value: string; label: string; data: AccessModel }>>([]);
  useEffect(() => {
    const filteredItems = filter != null ? accesses.filter(filter) : accesses;
    setOptions(
      filteredItems.map((item) => ({
        key: item.id,
        value: item.id,
        label: item.name,
        data: item,
      }))
    );
  }, [accesses, filter]);

  const renderOption = (key: string) => {
    const access = accesses.find((e) => e.id === key);
    if (!access) {
      return (
        <div className="flex items-center gap-2 overflow-hidden truncate">
          <Avatar shape="square" size="small" />
          <Typography.Text ellipsis>{key}</Typography.Text>
        </div>
      );
    }

    const provider = accessProvidersMap.get(access.provider);
    return (
      <div className="flex items-center gap-2 overflow-hidden truncate">
        <Avatar shape="square" src={provider?.icon} size="small" />
        <Typography.Text ellipsis>{access.name}</Typography.Text>
      </div>
    );
  };

  return (
    <Select
      {...props}
      filterOption={(inputValue, option) => {
        if (!option) return false;

        const value = inputValue.toLowerCase();
        return option.label.toLowerCase().includes(value);
      }}
      labelRender={({ value }) => {
        if (value != null) {
          return renderOption(value as string);
        }

        return <span style={{ color: themeToken.colorTextPlaceholder }}>{props.placeholder}</span>;
      }}
      loading={!loadedAtOnce}
      options={options}
      optionFilterProp="label"
      optionLabelProp={undefined}
      optionRender={(option) => renderOption(option.data.value)}
    />
  );
};

export default AccessSelect;
