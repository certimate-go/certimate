import { memo, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Card, Col, Empty, Flex, Input, type InputRef, Row, Typography } from "antd";

import Show from "@/components/Show";
import { type ACMEDns01Provider, acmeDns01ProvidersMap } from "@/domain/provider";

export type ACMEDns01ProviderPickerProps = {
  className?: string;
  style?: React.CSSProperties;
  autoFocus?: boolean;
  filter?: (record: ACMEDns01Provider) => boolean;
  placeholder?: string;
  onSelect?: (value: string) => void;
};

const ACMEDns01ProviderPicker = ({ className, style, autoFocus, filter, placeholder, onSelect }: ACMEDns01ProviderPickerProps) => {
  const { t } = useTranslation();

  const [keyword, setKeyword] = useState<string>();
  const keywordInputRef = useRef<InputRef>(null);
  useEffect(() => {
    if (autoFocus) {
      setTimeout(() => keywordInputRef.current?.focus(), 1);
    }
  }, []);

  const providers = useMemo(() => {
    return Array.from(acmeDns01ProvidersMap.values())
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
                <Col key={index} xs={24} md={12} span={12}>
                  <Card
                    className="h-16 w-full overflow-hidden shadow-xs"
                    styles={{ body: { height: "100%", padding: "0.5rem 1rem" } }}
                    hoverable
                    onClick={() => {
                      handleProviderTypeSelect(provider.type);
                    }}
                  >
                    <Flex className="size-full overflow-hidden" align="center" gap={8}>
                      <Avatar shape="square" src={provider.icon} size="small" />
                      <Typography.Text className="line-clamp-2 flex-1">{t(provider.name)}</Typography.Text>
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

export default memo(ACMEDns01ProviderPicker);
