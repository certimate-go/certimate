import { useMemo, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { IconCircleCheck, IconCircleX } from "@tabler/icons-react";
import { App, Button, Form, Input, List, Radio, Space, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { type MatrixVerifyStep, sendMatrixTestMessage, verifyMatrixConnection } from "@/api/matrix";
import { unwrapErrMsg } from "@/utils/error";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_MODES = [
  { value: "token", labelKey: "access.form.matrix_auth_mode.option.token" },
  { value: "password", labelKey: "access.form.matrix_auth_mode.option.password" },
] as const;

const STEP_LABEL_KEYS: Record<string, string> = {
  homeserver: "access.form.matrix_verify.step.homeserver",
  auth: "access.form.matrix_verify.step.auth",
  room: "access.form.matrix_verify.step.room",
};

const STEP_ERR_KEYS: Record<string, string> = {
  M_LIMIT_EXCEEDED: "access.form.matrix_verify.err.M_LIMIT_EXCEEDED",
  M_FORBIDDEN: "access.form.matrix_verify.err.M_FORBIDDEN",
  M_USER_DEACTIVATED: "access.form.matrix_verify.err.M_USER_DEACTIVATED",
  M_UIA_TIMEOUT: "access.form.matrix_verify.err.M_UIA_TIMEOUT",
};

const formatStepMessage = (t: (key: string, opts?: { seconds: number }) => string, step: MatrixVerifyStep): string => {
  if (step.code === "M_LIMIT_EXCEEDED") {
    if (step.retryAfterSec != null && step.retryAfterSec > 0) {
      return t("access.form.matrix_verify.err.M_LIMIT_EXCEEDED", { seconds: step.retryAfterSec });
    }
    return t("access.form.matrix_verify.err.M_LIMIT_EXCEEDED_generic");
  }
  if (step.code && STEP_ERR_KEYS[step.code]) {
    return t(STEP_ERR_KEYS[step.code]);
  }
  return step.message;
};

const AccessConfigFormFieldsProviderMatrix = () => {
  const { i18n, t } = useTranslation();
  const form = Form.useFormInstance();
  const { message, notification } = App.useApp();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  const authMode = Form.useWatch([parentNamePath, "authMode"], form) ?? "token";
  const sessionToken = Form.useWatch([parentNamePath, "sessionAccessToken"], form) as string | undefined;
  const accessId = Form.useWatch("id", form) as string | undefined;
  const homeserverUrl = Form.useWatch([parentNamePath, "homeserverUrl"], form) as string | undefined;
  const accessToken = Form.useWatch([parentNamePath, "accessToken"], form) as string | undefined;
  const userId = Form.useWatch([parentNamePath, "userId"], form) as string | undefined;
  const password = Form.useWatch([parentNamePath, "password"], form) as string | undefined;
  const roomId = Form.useWatch([parentNamePath, "roomId"], form) as string | undefined;

  const [verifying, setVerifying] = useState(false);
  const [sendingTest, setSendingTest] = useState(false);
  const [verifiedCredentialKey, setVerifiedCredentialKey] = useState<string | null>(null);
  const [verifySteps, setVerifySteps] = useState<MatrixVerifyStep[] | null>(null);

  const credentialKey = useMemo(
    () => [homeserverUrl ?? "", authMode, accessToken ?? "", userId ?? "", password ?? ""].join("\0"),
    [homeserverUrl, authMode, accessToken, userId, password]
  );
  const verifyOk = verifiedCredentialKey !== null && verifiedCredentialKey === credentialKey;

  const canSendTest = verifyOk && !!roomId?.trim();

  const handleVerifyConnection = async () => {
    setVerifySteps(null);
    setVerifiedCredentialKey(null);

    try {
      await form.validateFields([parentNamePath]);
    } catch {
      message.warning(t("common.errmsg.form_invalid"));
      return;
    }

    const config = form.getFieldValue(parentNamePath) as Record<string, unknown>;

    setVerifying(true);
    try {
      const result = await verifyMatrixConnection({ config, accessId });
      setVerifySteps(result.steps ?? []);
      if (result.ok) {
        setVerifiedCredentialKey(credentialKey);
        if (result.sessionAccessToken) {
          form.setFieldValue([parentNamePath, "sessionAccessToken"], result.sessionAccessToken);
          if (result.sessionDeviceId) {
            form.setFieldValue([parentNamePath, "sessionDeviceId"], result.sessionDeviceId);
          }
        }
        const extra = result.userId ? ` (${result.userId})` : "";
        if (result.sessionSaved) {
          message.success(t("access.form.matrix_session.saved_auto") + extra);
        } else if (result.sessionAccessToken) {
          message.success(t("access.form.matrix_session.saved_form") + extra);
        } else {
          message.success(t("access.form.matrix_verify.success") + extra);
        }
      } else {
        message.warning(t("access.form.matrix_verify.failed"));
      }
    } catch (err) {
      notification.error({
        title: t("common.text.request_error"),
        description: unwrapErrMsg(err),
      });
    } finally {
      setVerifying(false);
    }
  };

  const handleSendTestMessage = async () => {
    if (!canSendTest) {
      return;
    }

    try {
      await form.validateFields([parentNamePath]);
    } catch {
      message.warning(t("common.errmsg.form_invalid"));
      return;
    }

    const config = form.getFieldValue(parentNamePath) as Record<string, unknown>;

    setSendingTest(true);
    try {
      const result = await sendMatrixTestMessage({
        config,
        accessId,
        subject: t("access.form.matrix_test_send.subject"),
        message: t("access.form.matrix_test_send.body"),
      });

      if (result.sessionAccessToken) {
        form.setFieldValue([parentNamePath, "sessionAccessToken"], result.sessionAccessToken);
        if (result.sessionDeviceId) {
          form.setFieldValue([parentNamePath, "sessionDeviceId"], result.sessionDeviceId);
        }
      }

      if (result.sessionSaved) {
        message.success(t("access.form.matrix_session.saved_auto"));
      } else {
        message.success(t("access.form.matrix_test_send.success"));
      }
    } catch (err) {
      notification.error({
        title: t("common.text.request_error"),
        description: unwrapErrMsg(err),
      });
    } finally {
      setSendingTest(false);
    }
  };

  return (
    <>
      <Form.Item
        name={[parentNamePath, "homeserverUrl"]}
        label={t("access.form.matrix_homeserver_url.label")}
        rules={[formRule]}
        tooltip={{
          title: <span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_homeserver_url.tooltip") }} />,
        }}
        initialValue={initialValues?.homeserverUrl}
      >
        <Input type="url" placeholder={t("access.form.matrix_homeserver_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "authMode"]}
        label={t("access.form.matrix_auth_mode.label")}
        rules={[formRule]}
        initialValue={initialValues?.authMode ?? "token"}
      >
        <Radio.Group>
          {AUTH_MODES.map((o) => (
            <Radio key={o.value} value={o.value}>
              {t(o.labelKey)}
            </Radio>
          ))}
        </Radio.Group>
      </Form.Item>

      {authMode === "token" ? (
        <Form.Item
          name={[parentNamePath, "accessToken"]}
          label={t("access.form.matrix_access_token.label")}
          rules={[formRule]}
          help={t("access.form.matrix_access_token.help")}
          initialValue={initialValues?.accessToken}
        >
          <Input.Password placeholder={t("access.form.matrix_access_token.placeholder")} />
        </Form.Item>
      ) : (
        <>
          <Form.Item
            name={[parentNamePath, "userId"]}
            label={t("access.form.matrix_user_id.label")}
            rules={[formRule]}
            help={t("access.form.matrix_user_id.help")}
            initialValue={initialValues?.userId}
          >
            <Input placeholder={t("access.form.matrix_user_id.placeholder")} />
          </Form.Item>
          <Form.Item
            name={[parentNamePath, "password"]}
            label={t("access.form.matrix_password.label")}
            rules={[formRule]}
            help={t("access.form.matrix_password.help")}
            initialValue={initialValues?.password}
          >
            <Input.Password placeholder={t("access.form.matrix_password.placeholder")} />
          </Form.Item>
          <Form.Item name={[parentNamePath, "sessionAccessToken"]} hidden>
            <Input />
          </Form.Item>
          <Form.Item name={[parentNamePath, "sessionDeviceId"]} hidden>
            <Input />
          </Form.Item>
          {sessionToken ? <Typography.Text type="success">{t("access.form.matrix_session.active")}</Typography.Text> : null}
        </>
      )}

      <Form.Item
        name={[parentNamePath, "roomId"]}
        label={t("access.form.matrix_room_id.label")}
        rules={[formRule]}
        extra={<span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_room_id.help") }} />}
        tooltip={{
          title: <span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_room_id.tooltip") }} />,
        }}
        initialValue={initialValues?.roomId}
      >
        <Input allowClear placeholder={t("access.form.matrix_room_id.placeholder")} />
      </Form.Item>

      <Form.Item>
        <Space direction="vertical" style={{ width: "100%" }}>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 0 }}>
            <span dangerouslySetInnerHTML={{ __html: t("access.form.matrix_verify.hint") }} />
          </Typography.Paragraph>
          <Space wrap>
            <Button loading={verifying} type="default" onClick={handleVerifyConnection}>
              {t("access.form.matrix_verify.button")}
            </Button>
            {canSendTest ? (
              <Button loading={sendingTest} type="primary" onClick={handleSendTestMessage}>
                {t("access.form.matrix_test_send.button")}
              </Button>
            ) : null}
          </Space>
          {canSendTest ? (
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {t("access.form.matrix_test_send.hint")}
            </Typography.Text>
          ) : null}
          {verifySteps != null && verifySteps.length > 0 ? (
            <List
              size="small"
              dataSource={verifySteps}
              renderItem={(step) => (
                <List.Item style={{ paddingInline: 0 }}>
                  <Space align="start">
                    {step.ok ? <IconCircleCheck color="var(--ant-color-success)" size={18} /> : <IconCircleX color="var(--ant-color-error)" size={18} />}
                    <Space direction="vertical" size={0}>
                      <Typography.Text strong>{t(STEP_LABEL_KEYS[step.name] ?? step.name)}</Typography.Text>
                      <Typography.Text type={step.ok ? "secondary" : "danger"}>{formatStepMessage(t, step)}</Typography.Text>
                      {step.detail ? (
                        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                          {step.detail}
                        </Typography.Text>
                      ) : null}
                    </Space>
                  </Space>
                </List.Item>
              )}
            />
          ) : null}
        </Space>
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<Record<string, unknown>> => {
  return {
    authMode: "token",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      homeserverUrl: z.url(t("common.errmsg.url_invalid")),
      authMode: z.enum(["token", "password"]),
      accessToken: z.string().optional(),
      userId: z.string().optional(),
      password: z.string().optional(),
      roomId: z.string().optional(),
    })
    .superRefine((data, ctx) => {
      if (data.authMode === "token" && !data.accessToken?.trim()) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: t("access.form.matrix_access_token.placeholder"),
          path: ["accessToken"],
        });
      }
      if (data.authMode === "password") {
        if (!data.userId?.trim()) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: t("access.form.matrix_user_id.placeholder"),
            path: ["userId"],
          });
        }
        if (!data.password?.trim()) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: t("access.form.matrix_password.placeholder"),
            path: ["password"],
          });
        }
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderMatrix, {
  getInitialValues,
  getSchema,
});

export default _default;
