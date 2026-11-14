import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Alert, Button, Form, type FormInstance, Space, Tag, Typography } from "antd";
import { z } from "zod";

import { checkDNSPersist, prepareDNSPersist } from "@/api/acmeAccounts";
import { type DNSPersistCheck, type DNSPersistPreparation } from "@/domain/acmeAccount";
import { unwrapErrMsg } from "@/utils/error";

type SetupState =
  | { kind: "idle" }
  | { kind: "preparing" }
  | { kind: "failed"; message: string }
  | { kind: "prepared"; preparation: DNSPersistPreparation }
  | { kind: "checking"; preparation: DNSPersistPreparation }
  | { kind: "checked"; preparation: DNSPersistPreparation; result: DNSPersistCheck }
  | { kind: "check_failed"; preparation: DNSPersistPreparation; message: string };

const preparationFields = ["domains", "contactEmail", "caProvider", "caProviderAccessId", "caProviderConfig", "keyAlgorithm", "acmeProfile"];
const identityFields = ["identifier", "challengeType", ...preparationFields];
const splitValues = (value: string) =>
  value
    .split(";")
    .map((part) => part.trim())
    .filter(Boolean);
const optionalText = z
  .string()
  .nullish()
  .transform((value) => value || undefined);
const payloadSchema = z.object({
  domains: z.string().transform(splitValues),
  contactEmail: z.email(),
  caProvider: optionalText,
  caProviderAccessId: optionalText,
  caProviderConfig: z
    .record(z.string(), z.unknown())
    .nullish()
    .transform((value) => value ?? undefined),
  keyAlgorithm: optionalText,
  acmeProfile: optionalText,
  nameservers: z
    .string()
    .nullish()
    .transform((value) => splitValues(value || "")),
});
const identityOf = (values: Record<string, unknown>) => JSON.stringify(identityFields.map((field) => values[field]));
const nameserversOf = (value: unknown) => splitValues(typeof value === "string" ? value : "");

export default function DNSPersistSetup({ enabled, form, value, id }: { enabled: boolean; form: FormInstance; value?: string | null; id?: string }) {
  const { t } = useTranslation();
  const [state, setState] = useState<SetupState>({ kind: "idle" });
  const identity = Form.useWatch(identityOf, { form, preserve: true });
  const nameservers = Form.useWatch((values) => JSON.stringify(nameserversOf(values.nameservers)), { form, preserve: true });
  const previousIdentity = useRef(form.getFieldValue("identifier") ? identityOf(form.getFieldsValue(true)) : undefined);
  const previousNameservers = useRef(JSON.stringify(nameserversOf(form.getFieldValue("nameservers"))));
  const generation = useRef(0);
  const checkGeneration = useRef(0);

  useEffect(() => {
    if (identity === undefined || !form.getFieldValue("identifier") || identity === previousIdentity.current) return;
    const initialized = previousIdentity.current !== undefined;
    previousIdentity.current = identity;
    if (!initialized) return;
    generation.current += 1;
    checkGeneration.current += 1;
    form.setFieldValue("dnsPersistAccountId", undefined);
    setState({ kind: "idle" });
  }, [form, identity]);

  useEffect(() => {
    if (nameservers === undefined || nameservers === previousNameservers.current) return;
    previousNameservers.current = nameservers;
    checkGeneration.current += 1;
    setState((current) => ("preparation" in current ? { kind: "prepared", preparation: current.preparation } : current));
  }, [nameservers]);

  useEffect(
    () => () => {
      generation.current += 1;
      checkGeneration.current += 1;
    },
    []
  );

  const handlePrepare = async () => {
    const requestedIdentity = identityOf(form.getFieldsValue(true));
    const requestGeneration = ++generation.current;
    const isCurrent = () => requestGeneration === generation.current && requestedIdentity === identityOf(form.getFieldsValue(true));
    try {
      await form.validateFields([...preparationFields, "nameservers"]);
    } catch {
      return;
    }
    if (!isCurrent()) return;
    form.setFieldValue("dnsPersistAccountId", undefined);
    setState({ kind: "preparing" });
    try {
      const response = await prepareDNSPersist(payloadSchema.parse(form.getFieldsValue(true)));
      if (!isCurrent()) return;
      form.setFieldValue("dnsPersistAccountId", response.data.accountId);
      setState({ kind: "prepared", preparation: response.data });
    } catch (error) {
      if (isCurrent()) setState({ kind: "failed", message: unwrapErrMsg(error) });
    }
  };

  const handleCheck = async () => {
    if (!("preparation" in state)) return;
    const preparation = state.preparation;
    const requestedIdentity = identityOf(form.getFieldsValue(true));
    const requestedNameservers = JSON.stringify(nameserversOf(form.getFieldValue("nameservers")));
    const requestGeneration = generation.current;
    const requestCheckGeneration = ++checkGeneration.current;
    const isCurrent = () =>
      requestGeneration === generation.current &&
      requestCheckGeneration === checkGeneration.current &&
      requestedIdentity === identityOf(form.getFieldsValue(true)) &&
      requestedNameservers === JSON.stringify(nameserversOf(form.getFieldValue("nameservers")));
    try {
      await form.validateFields(["nameservers"]);
    } catch {
      return;
    }
    if (!isCurrent()) return;
    setState({ kind: "checking", preparation });
    try {
      const response = await checkDNSPersist({ records: preparation.records, nameservers: nameserversOf(form.getFieldValue("nameservers")) });
      if (isCurrent()) setState({ kind: "checked", preparation, result: response.data });
    } catch (error) {
      if (isCurrent()) setState({ kind: "check_failed", preparation, message: unwrapErrMsg(error) });
    }
  };

  if (!enabled) return null;
  const preparation = "preparation" in state ? state.preparation : undefined;
  const busy = state.kind === "preparing" || state.kind === "checking";
  return (
    <div id={id} className="flex flex-col gap-3">
      <Alert
        showIcon
        type="warning"
        title={t("workflow_node.apply.form.dns_persist.label")}
        description={t("workflow_node.apply.form.dns_persist.description")}
      />
      <Typography.Text>{t("workflow_node.apply.form.dns_persist.instructions")}</Typography.Text>
      {!!value && !preparation && <Typography.Text type="secondary">{t("workflow_node.apply.form.dns_persist.saved_binding")}</Typography.Text>}
      <Space wrap>
        <Button aria-label={t("workflow_node.apply.form.dns_persist.generate")} disabled={busy} loading={state.kind === "preparing"} onClick={handlePrepare}>
          {t("workflow_node.apply.form.dns_persist.generate")}
        </Button>
        <Button
          aria-label={t("workflow_node.apply.form.dns_persist.check")}
          disabled={!preparation || busy}
          loading={state.kind === "checking"}
          onClick={handleCheck}
        >
          {t("workflow_node.apply.form.dns_persist.check")}
        </Button>
      </Space>
      {(state.kind === "failed" || state.kind === "check_failed") && (
        <Alert showIcon type="error" title={t("workflow_node.apply.form.dns_persist.failed")} description={state.message} />
      )}
      {preparation && (
        <>
          <div>
            <Typography.Text type="secondary">{t("workflow_node.apply.form.dns_persist.account")}</Typography.Text>
            <Typography.Paragraph className="mb-0 break-all" copyable>
              {preparation.accountUrl}
            </Typography.Paragraph>
          </div>
          {preparation.records.map((record) => {
            const result = state.kind === "checked" ? state.result.records.find((item) => item.name === record.name && item.value === record.value) : undefined;
            return (
              <div key={`${record.name}:${record.value}`} className="rounded-lg border border-solid border-gray-200 p-3 dark:border-gray-700">
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <Typography.Text strong>{record.domain}</Typography.Text>
                  <Tag>TXT</Tag>
                  {result && (
                    <Tag color={result.status === "ready" ? "success" : "warning"}>{t(`workflow_node.apply.form.dns_persist.status.${result.status}`)}</Tag>
                  )}
                </div>
                <Typography.Text type="secondary">{t("workflow_node.apply.form.dns_persist.name")}</Typography.Text>
                <Typography.Paragraph className="break-all" copyable>
                  {record.name}
                </Typography.Paragraph>
                <Typography.Text type="secondary">{t("workflow_node.apply.form.dns_persist.value")}</Typography.Text>
                <Typography.Paragraph className="mb-0 break-all" copyable>
                  {record.value}
                </Typography.Paragraph>
                {result?.message && <Typography.Text type="secondary">{result.message}</Typography.Text>}
              </div>
            );
          })}
        </>
      )}
      {state.kind === "checked" && (
        <Alert
          showIcon
          type={state.result.ready ? "success" : "warning"}
          title={t(`workflow_node.apply.form.dns_persist.${state.result.ready ? "ready" : "not_ready"}`)}
        />
      )}
    </div>
  );
}
