import { memo, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Card, Col, Empty, Flex, Input, type InputRef, Row, Tag, Typography } from "antd";

import Show from "@/components/Show";
import { ACCESS_USAGES, type AccessProvider, type AccessUsageType, accessProvidersMap } from "@/domain/provider";
import { mergeCls } from "@/utils/css";

export type AccessProviderPickerProps = {
  className?: string;
  style?: React.CSSProperties;
  autoFocus?: boolean;
  filter?: (record: AccessProvider) => boolean;
  placeholder?: string;
  showOptionTags?: boolean | { [key in AccessUsageType]?: boolean };
  onSelect?: (value: string) => void;
};

const AccessProviderPicker = ({ className, style, autoFocus, filter, placeholder, showOptionTags, onSelect }: AccessProviderPickerProps) => {
  const { t } = useTranslation();

  const [keyword, setKeyword] = useState<string>();
  const keywordInputRef = useRef<InputRef>(null);
  useEffect(() => {
    if (autoFocus) {
      setTimeout(() => keywordInputRef.current?.focus(), 1);
    }
  }, []);

  const providers = useMemo(() => {
    return Array.from(accessProvidersMap.values())
      .filter((provider) => {
        if (filter) {
          return filter(provider);
        }

        return true;
      })
      .filter((provider) => {
        if (keyword) {
          const value = keyword.toLowerCase();
          return provider.type.toLowerCase().includes(value) || t(provider.name).toLowerCase().includes(value);
        }

        return true;
      });
  }, [filter, keyword]);

  const showOptionTagForDNS = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags[ACCESS_USAGES.DNS] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForHosting = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags[ACCESS_USAGES.HOSTING] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForCA = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags[ACCESS_USAGES.CA] : !!showOptionTags;
  }, [showOptionTags]);
  const showOptionTagForNotification = useMemo(() => {
    return typeof showOptionTags === "object" ? !!showOptionTags[ACCESS_USAGES.NOTIFICATION] : !!showOptionTags;
  }, [showOptionTags]);

  const handleProviderTypeSelect = (value: string) => {
    onSelect?.(value);
  };

  return (
    <div className={className} style={style}>
      <Input.Search ref={keywordInputRef} placeholder={placeholder ?? t("common.text.search")} onChange={(e) => setKeyword(e.target.value.trim())} />

      <div className="mt-4">
        <Show when={providers.length > 0} fallback={<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}>
          <Row gutter={[16, 16]}>
            {providers.map((provider, index) => {
              return (
                <Col key={index} xs={24} md={12} span={8}>
                  <Card
                    className={mergeCls("h-20 w-full overflow-hidden shadow-xs", provider.builtin ? " cursor-not-allowed" : "")}
                    styles={{ body: { height: "100%", padding: "0.5rem 1rem" } }}
                    hoverable
                    onClick={() => {
                      if (provider.builtin) {
                        return;
                      }

                      handleProviderTypeSelect(provider.type);
                    }}
                  >
                    <Flex className="size-full overflow-hidden" align="center" gap={8}>
                      <Avatar shape="square" src={provider.icon} size="small" />
                      <div className="flex-1 overflow-hidden">
                        <Typography.Text className="mb-1 line-clamp-1" type={provider.builtin ? "secondary" : undefined}>
                          {t(provider.name)}
                        </Typography.Text>
                        <div className="origin-left scale-75">
                          <Show when={provider.builtin}>
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
                    </Flex>
                  </Card>
                </Col>
              );
            })}
          </Row>
        </Show>
      </div>
    </div>
  );
};

export default memo(AccessProviderPicker);
