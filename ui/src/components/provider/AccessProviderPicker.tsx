import { useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMount } from "ahooks";
import { Avatar, Card, Empty, Input, type InputRef, Tag, Typography } from "antd";

import Show from "@/components/Show";
import { ACCESS_USAGES, type AccessProvider, type AccessUsageType, accessProvidersMap } from "@/domain/provider";
import { mergeCls } from "@/utils/css";

import { type SharedPickerProps, usePickerDataSource, usePickerWrapperCols } from "./_shared";

export interface AccessProviderPickerProps extends SharedPickerProps<AccessProvider> {
  showOptionTags?: boolean | { [key in AccessUsageType | "builtin"]?: boolean };
}

const AccessProviderPicker = ({
  className,
  style,
  autoFocus,
  gap = "middle",
  placeholder,
  showOptionTags,
  showSearch = false,
  onFilter,
  onSelect,
}: AccessProviderPickerProps) => {
  const { t } = useTranslation();

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
  const showOptionTagAnyhow = useMemo(() => {
    return showOptionTagForDNS || showOptionTagForHosting || showOptionTagForCA || showOptionTagForNotification || showOptionTagForBuiltin;
  }, [showOptionTagForDNS, showOptionTagForHosting, showOptionTagForCA, showOptionTagForNotification, showOptionTagForBuiltin]);

  const { wrapperElRef, cols } = usePickerWrapperCols(showOptionTagAnyhow ? 240 : 200);

  const [keyword, setKeyword] = useState<string>();
  const keywordInputRef = useRef<InputRef>(null);
  useMount(() => {
    if (autoFocus) {
      setTimeout(() => keywordInputRef.current?.focus(), 1);
    }
  });

  const dataSources = usePickerDataSource({
    dataSource: Array.from(accessProvidersMap.values()),
    filters: [onFilter!],
    keyword: keyword,
  });

  const renderOption = (provider: AccessProvider) => {
    return (
      <div className="group/provider size-full" key={provider.type}>
        <Card
          className={mergeCls("size-full overflow-hidden shadow", provider.builtin ? "cursor-not-allowed" : void 0)}
          styles={{
            body: {
              height: "100%",
              padding: "1.25rem 1rem",
            },
          }}
          hoverable
          onClick={() => {
            if (provider.builtin) {
              return;
            }

            handleProviderTypeSelect(provider.type);
          }}
        >
          <div className="flex size-full flex-col">
            <div className="flex flex-1 justify-between gap-3">
              <div className="flex-1">
                <Typography.Text type={provider.builtin ? "secondary" : void 0}>{t(provider.name) || "\u00A0"}</Typography.Text>
              </div>
              <div className="transition-all group-hover/provider:scale-110">
                <Avatar className="bg-stone-50" icon={<img src={provider.icon} />} shape="square" size={28} />
              </div>
            </div>
            <Show when={showOptionTagAnyhow}>
              <div className="flex origin-left scale-80 items-center gap-1 whitespace-nowrap">
                <Show when={showOptionTagForBuiltin && provider.builtin}>
                  <Tag className="mt-4 -mb-2" color="default">
                    {t("access.props.provider.builtin")}
                  </Tag>
                </Show>
                <Show when={showOptionTagForDNS && provider.usages.includes(ACCESS_USAGES.DNS)}>
                  <Tag className="mt-4 -mb-2" color="#d93f0b">
                    {t("access.props.provider.usage.dns")}
                  </Tag>
                </Show>
                <Show when={showOptionTagForHosting && provider.usages.includes(ACCESS_USAGES.HOSTING)}>
                  <Tag className="mt-4 -mb-2" color="#0052cc">
                    {t("access.props.provider.usage.hosting")}
                  </Tag>
                </Show>
                <Show when={showOptionTagForCA && provider.usages.includes(ACCESS_USAGES.CA)}>
                  <Tag className="mt-4 -mb-2" color="#0e8a16">
                    {t("access.props.provider.usage.ca")}
                  </Tag>
                </Show>
                <Show when={showOptionTagForNotification && provider.usages.includes(ACCESS_USAGES.NOTIFICATION)}>
                  <Tag className="mt-4 -mb-2" color="#1d76db">
                    {t("access.props.provider.usage.notification")}
                  </Tag>
                </Show>
              </div>
            </Show>
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

      <Show when={dataSources.filtered.length > 0} fallback={<Empty description={t("provider.text.nodata")} image={Empty.PRESENTED_IMAGE_SIMPLE} />}>
        <div
          className={mergeCls("grid w-full gap-2", `grid-cols-${cols}`, {
            "gap-4": gap === "large",
            "gap-2": gap === "middle",
            "gap-1": gap === "small",
            [`gap-${+gap || "2"}`]: typeof gap === "number",
          })}
        >
          {dataSources.filtered.map((provider) => renderOption(provider))}
        </div>
      </Show>
    </div>
  );
};

export default AccessProviderPicker;
