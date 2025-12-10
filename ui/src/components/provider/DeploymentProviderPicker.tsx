import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMount } from "ahooks";
import { Avatar, Card, Divider, Empty, Flex, Input, type InputRef, Tabs, Tooltip, Typography } from "antd";

import Show from "@/components/Show";
import { DEPLOYMENT_CATEGORIES, type DeploymentProvider, deploymentProvidersMap } from "@/domain/provider";
import { mergeCls } from "@/utils/css";

import { type SharedPickerProps, usePickerDataSource, usePickerWrapperCols } from "./_shared";

export interface DeploymentProviderPickerProps extends SharedPickerProps<DeploymentProvider> {
  showAvailability?: boolean;
}

const DeploymentProviderPicker = ({
  className,
  style,
  autoFocus,
  gap = "middle",
  placeholder,
  showAvailability = false,
  showSearch = false,
  onFilter,
  onSelect,
}: DeploymentProviderPickerProps) => {
  const { t } = useTranslation();

  const { wrapperElRef, cols } = usePickerWrapperCols(320);

  const [category, setCategory] = useState<string>(DEPLOYMENT_CATEGORIES.ALL);

  const [keyword, setKeyword] = useState<string>();
  const keywordInputRef = useRef<InputRef>(null);
  useMount(() => {
    if (autoFocus) {
      setTimeout(() => keywordInputRef.current?.focus(), 1);
    }
  });

  const dataSources = usePickerDataSource({
    dataSource: Array.from(deploymentProvidersMap.values()),
    filters: [
      onFilter!,
      (_, provider) => {
        if (category && category !== DEPLOYMENT_CATEGORIES.ALL) {
          return provider.category === category;
        }

        return true;
      },
    ],
    keyword: keyword,
    deps: [category],
  });

  const renderOption = (provider: DeploymentProvider, transparent: boolean = false) => {
    return (
      <div key={provider.type}>
        <Card
          className="group/provider h-16 w-full overflow-hidden shadow"
          styles={{ body: { height: "100%", padding: "0.5rem 1rem" } }}
          hoverable
          onClick={() => {
            handleProviderTypeSelect(provider.type);
          }}
        >
          <div className={mergeCls("size-full", transparent ? "transition-opacity opacity-75 group-hover/provider:opacity-100" : void 0)}>
            <div className="flex size-full items-center gap-4 overflow-hidden">
              <div>
                <Avatar className="bg-stone-50" icon={<img src={provider.icon} />} shape="square" size={28} />
              </div>
              <div className="flex-1 overflow-hidden">
                <div className="line-clamp-2 max-w-full">
                  <Tooltip title={t(provider.name)} mouseEnterDelay={1}>
                    <Typography.Text>{t(provider.name) || "\u00A0"}</Typography.Text>
                  </Tooltip>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </div>
    );
  };

  const handleProviderTypeSelect = (value: string) => {
    onSelect?.(value);
  };

  return (
    <div className={className} style={style} ref={wrapperElRef}>
      <Show when={showSearch}>
        <div className="mb-4">
          <Input.Search ref={keywordInputRef} placeholder={placeholder ?? t("common.text.search")} onChange={(e) => setKeyword(e.target.value.trim())} />
        </div>
      </Show>

      <Flex>
        <Tabs
          defaultActiveKey={DEPLOYMENT_CATEGORIES.ALL}
          items={[
            DEPLOYMENT_CATEGORIES.ALL,
            DEPLOYMENT_CATEGORIES.CDN,
            DEPLOYMENT_CATEGORIES.STORAGE,
            DEPLOYMENT_CATEGORIES.LOADBALANCE,
            DEPLOYMENT_CATEGORIES.FIREWALL,
            DEPLOYMENT_CATEGORIES.AV,
            DEPLOYMENT_CATEGORIES.ACCELERATOR,
            DEPLOYMENT_CATEGORIES.APIGATEWAY,
            DEPLOYMENT_CATEGORIES.SERVERLESS,
            DEPLOYMENT_CATEGORIES.WEBSITE,
            DEPLOYMENT_CATEGORIES.SSL,
            DEPLOYMENT_CATEGORIES.OTHER,
          ].map((key) => ({
            key: key,
            label: t(`provider.category.${key}`),
          }))}
          size="small"
          tabBarStyle={{ marginLeft: "-1rem" }}
          tabPlacement="start"
          onChange={(key) => setCategory(key)}
        />

        <div className="flex-1">
          <Show when={dataSources.filtered.length > 0} fallback={<Empty description={t("provider.text.nodata")} image={Empty.PRESENTED_IMAGE_SIMPLE} />}>
            <div
              className={mergeCls("grid w-full gap-2", `grid-cols-${cols}`, {
                "gap-4": gap === "large",
                "gap-2": gap === "middle",
                "gap-1": gap === "small",
                [`gap-${+gap || "2"}`]: typeof gap === "number",
              })}
            >
              {(showAvailability ? dataSources.available : dataSources.filtered).map((provider) => renderOption(provider))}
            </div>

            <Show when={showAvailability && dataSources.unavailable.length > 0}>
              <Divider size="small">
                <Typography.Text className="text-xs font-normal" type="secondary">
                  {t("provider.text.unavailable_divider")}
                </Typography.Text>
              </Divider>

              <div
                className={mergeCls("grid w-full gap-2", `grid-cols-${cols}`, {
                  "gap-4": gap === "large",
                  "gap-2": gap === "middle",
                  "gap-1": gap === "small",
                  [`gap-${+gap || "2"}`]: typeof gap === "number",
                })}
              >
                {dataSources.unavailable.map((provider) => renderOption(provider, true))}
              </div>
            </Show>
          </Show>
        </div>
      </Flex>
    </div>
  );
};

export default DeploymentProviderPicker;
