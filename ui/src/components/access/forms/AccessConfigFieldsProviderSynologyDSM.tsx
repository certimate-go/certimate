import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFieldsProviderSynologyDSM = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item
        name={[parentNamePath, "serverUrl"]}
        initialValue={initialValues.serverUrl}
        label={t("access.form.synologydsm_server_url.label")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.synologydsm_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "username"]}
        initialValue={initialValues.username}
        label={t("access.form.synologydsm_username.label")}
        rules={[formRule]}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.synologydsm_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "password"]}
        initialValue={initialValues.password}
        label={t("access.form.synologydsm_password.label")}
        rules={[formRule]}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.synologydsm_password.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "totpSecret"]}
        initialValue={initialValues.totpSecret}
        label={t("access.form.synologydsm_totp_secret.label")}
        extra={t("access.form.synologydsm_totp_secret.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.synologydsm_totp_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.synologydsm_totp_secret.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "allowInsecureConnections"]}
        initialValue={initialValues.allowInsecureConnections}
        rules={[formRule]}
        label={t("access.form.shared_allow_insecure_conns.label")}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    serverUrl: "http://<your-host-addr>:5000/",
    username: "",
    password: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    username: z.string().nonempty(t("access.form.synologydsm_username.placeholder")),
    password: z.string().nonempty(t("access.form.synologydsm_password.placeholder")),
    totpSecret: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return /^[A-Z2-7]{16,32}$/.test(v);
      }),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFieldsProviderSynologyDSM, {
  getInitialValues,
  getSchema,
});

export default _default;
