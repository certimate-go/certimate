import { memo, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Card, Col, Empty, Flex, Input, type InputRef, Row, Tabs, Tooltip, Typography } from "antd";

import Show from "@/components/Show";
import { DEPLOYMENT_CATEGORIES, type DeploymentProvider, deploymentProvidersMap } from "@/domain/provider";

export type DeploymentProviderPickerProps = {
  className?: string;
  style?: React.CSSProperties;
  autoFocus?: boolean;
  filter?: (record: DeploymentProvider) => boolean;
  placeholder?: string;
  onSelect?: (value: string) => void;
};

const DeploymentProviderPicker = ({ className, style, autoFocus, filter, placeholder, onSelect }: DeploymentProviderPickerProps) => {
  const { t } = useTranslation();

  const [category, setCategory] = useState<string>(DEPLOYMENT_CATEGORIES.ALL);

  const [keyword, setKeyword] = useState<string>();
  const keywordInputRef = useRef<InputRef>(null);
  useEffect(() => {
    if (autoFocus) {
      setTimeout(() => keywordInputRef.current?.focus(), 1);
    }
  }, []);

  const providers = useMemo(() => {
    return Array.from(deploymentProvidersMap.values())
      .filter((provider) => {
        if (filter) {
          return filter(provider);
        }

        return true;
      })
      .filter((provider) => {
        if (category && category !== DEPLOYMENT_CATEGORIES.ALL) {
          return provider.category === category;
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
  }, [filter, category, keyword]);

  const handleProviderTypeSelect = (value: string) => {
    onSelect?.(value);
  };

  return (
    <div className={className} style={style}>
      <Input.Search ref={keywordInputRef} placeholder={placeholder ?? t("common.text.search")} onChange={(e) => setKeyword(e.target.value.trim())} />

      <div className="mt-4">
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
              DEPLOYMENT_CATEGORIES.NAS,
              DEPLOYMENT_CATEGORIES.OTHER,
            ].map((key) => ({
              key: key,
              label: t(`provider.category.${key}`),
            }))}
            size="small"
            tabBarStyle={{ marginLeft: "-1rem" }}
            tabPosition="left"
            onChange={(key) => setCategory(key)}
          />

          <div className="flex-1">
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
                        <Tooltip title={t(provider.name)} mouseEnterDelay={1}>
                          <Flex className="size-full overflow-hidden" align="center" gap={8}>
                            <Avatar shape="square" src={provider.icon} size="small" />
                            <Typography.Text className="line-clamp-2 flex-1">{t(provider.name)}</Typography.Text>
                          </Flex>
                        </Tooltip>
                      </Card>
                    </Col>
                  );
                })}
              </Row>
            </Show>
          </div>
        </Flex>
      </div>
    </div>
  );
};

export default memo(DeploymentProviderPicker);
