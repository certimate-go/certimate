import { forwardRef, useImperativeHandle, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Card, Divider, Empty, Input, type InputRef, Tooltip, Typography } from "antd";

import Show from "@/components/Show";
import { type NotificationProvider, notificationProvidersMap } from "@/domain/provider";
import { mergeCls } from "@/utils/css";

import { type SharedPickerProps, usePickerDataSource, usePickerWrapperCols } from "./_shared";

export interface NotificationProviderPickerProps extends SharedPickerProps<NotificationProvider> {
  showAvailability?: boolean;
}

export interface NotificationProviderPickerInstance {
  inputRef: InputRef | null;
}

const NotificationProviderPicker = forwardRef<NotificationProviderPickerInstance, NotificationProviderPickerProps>(
  ({ className, style, gap = "medium", placeholder, showAvailability = false, showSearch = false, onFilter, onSelect }, ref) => {
    const { t } = useTranslation();

    const { wrapperElRef, cols } = usePickerWrapperCols(320);

    const [keyword, setKeyword] = useState<string>();
    const keywordInputRef = useRef<InputRef>(null);

    const dataSources = usePickerDataSource({
      dataSource: Array.from(notificationProvidersMap.values()),
      filters: [onFilter!],
      keyword: keyword,
    });

    const renderOption = (provider: NotificationProvider, transparent: boolean = false) => {
      return (
        <div key={provider.type}>
          <Card
            className="group/provider h-16 w-full overflow-hidden shadow-sm"
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

    useImperativeHandle(ref, () => ({
      get inputRef() {
        return keywordInputRef.current;
      },
    }));

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
              "gap-2": gap === "medium",
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
                "gap-2": gap === "medium",
                "gap-1": gap === "small",
                [`gap-${+gap || "2"}`]: typeof gap === "number",
              })}
            >
              {dataSources.unavailable.map((provider) => renderOption(provider, true))}
            </div>
          </Show>
        </Show>
      </div>
    );
  }
);

export default NotificationProviderPicker;
