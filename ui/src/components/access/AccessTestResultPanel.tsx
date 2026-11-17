import { useTranslation } from "react-i18next";
import { Alert, Flex, Tag, Typography } from "antd";

import { ACCESS_TEST_REASON_CODES, type AccessTestResult } from "@/api/accesses";

export interface AccessTestResultPanelProps {
  className?: string;
  result: AccessTestResult;
}

type AccessTestResultStatus = AccessTestResult["status"];
type AccessTestItemStatus = AccessTestResult["items"][number]["status"];

const ALERT_TYPES: Record<AccessTestResultStatus, "success" | "error" | "warning" | "info"> = {
  passed: "success",
  failed: "error",
  inconclusive: "warning",
  unsupported: "info",
  unavailable: "warning",
};

const ITEM_TAG_COLORS: Record<AccessTestItemStatus, string> = {
  passed: "success",
  failed: "error",
  skipped: "default",
};

const ITEM_LABEL_KEYS: Record<string, string> = {
  connect: "access.test.item.connect",
  auth: "access.test.item.auth",
};

const SCOPE_KEYS: Partial<Record<AccessTestResultStatus, string>> = {
  passed: "access.test.scope.passed",
  failed: "access.test.scope.failed",
  inconclusive: "access.test.scope.inconclusive",
};

const AccessTestResultPanel = ({ className, result }: AccessTestResultPanelProps) => {
  const { t } = useTranslation();

  const reasonText = !result.reasonCode
    ? null
    : (ACCESS_TEST_REASON_CODES as readonly string[]).includes(result.reasonCode)
      ? t(`access.test.reason.${result.reasonCode}`)
      : t("access.test.reason.unknown");

  const scopeKey = SCOPE_KEYS[result.status];

  return (
    <Alert
      className={className}
      type={ALERT_TYPES[result.status]}
      showIcon
      message={t(`access.test.status.${result.status}`)}
      description={
        <Flex vertical gap="small">
          {reasonText ? <div>{reasonText}</div> : null}
          {result.items.length > 0 ? (
            <Flex gap="small" wrap="wrap">
              {result.items.map((item) => (
                <Tag color={ITEM_TAG_COLORS[item.status]} key={item.id}>
                  {t("access.test.item.format", {
                    label: ITEM_LABEL_KEYS[item.id] ? t(ITEM_LABEL_KEYS[item.id]) : item.id,
                    status: t(`access.test.item_status.${item.status}`),
                  })}
                </Tag>
              ))}
            </Flex>
          ) : null}
          {result.diagnostics.length > 0 ? (
            <div>
              <Typography.Text type="secondary">{t("access.test.diagnostics.label")}</Typography.Text>
              {result.diagnostics.map((line, index) => (
                <div key={index}>{line}</div>
              ))}
            </div>
          ) : null}
          {scopeKey ? <Typography.Text type="secondary">{t(scopeKey)}</Typography.Text> : null}
        </Flex>
      }
    />
  );
};

export default AccessTestResultPanel;
