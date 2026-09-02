import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_METHOD_PASSWORD = "password" as const;
const AUTH_METHOD_APIKEY = "apikey" as const;

const AccessConfigFormFieldsProviderLeCDN = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldAuthMethod = Form.useWatch<string>([parentNamePath, "authMethod"], formInst);

  return (
    <>
      <Form.Item name={[parentNamePath, "serverUrl"]} initialValue={initialValues.serverUrl} label={t("access.form.lecdn_server_url.label")} rules={[formRule]}>
        <Input type="url" placeholder={t("access.form.lecdn_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiVersion"]}
        initialValue={initialValues.apiVersion}
        label={t("access.form.lecdn_api_version.label")}
        rules={[formRule]}
      >
        <Select options={["v3", "v4"].map((s) => ({ label: s, value: s }))} placeholder={t("access.form.lecdn_api_version.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "apiRole"]} initialValue={initialValues.apiRole} label={t("access.form.lecdn_api_role.label")} rules={[formRule]}>
        <Radio.Group
          options={["client", "master"].map((s) => ({
            label: t(`access.form.lecdn_api_role.option.${s}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "authMethod"]}
        initialValue={initialValues.authMethod}
        label={t("access.form.lecdn_auth_method.label")}
        rules={[formRule]}
      >
        <Radio.Group block>
          <Radio.Button value={AUTH_METHOD_PASSWORD}>{t("access.form.lecdn_auth_method.option.password.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_APIKEY}>{t("access.form.lecdn_auth_method.option.apikey.label")}</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_PASSWORD}>
        <Form.Item name={[parentNamePath, "username"]} initialValue={initialValues.username} label={t("access.form.lecdn_username.label")} rules={[formRule]}>
          <Input autoComplete="new-password" placeholder={t("access.form.lecdn_username.placeholder")} />
        </Form.Item>

        <Form.Item name={[parentNamePath, "password"]} initialValue={initialValues.password} label={t("access.form.lecdn_password.label")} rules={[formRule]}>
          <Input.Password autoComplete="new-password" placeholder={t("access.form.lecdn_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldAuthMethod === AUTH_METHOD_APIKEY}>
        <Form.Item name={[parentNamePath, "apiKey"]} initialValue={initialValues.apiKey} label={t("access.form.lecdn_api_key.label")} rules={[formRule]}>
          <Input.Password autoComplete="new-password" placeholder={t("access.form.lecdn_api_key.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "allowInsecureConnections"]}
        initialValue={initialValues.allowInsecureConnections}
        label={t("access.form.shared_allow_insecure_conns.label")}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    serverUrl: "http://<your-host-addr>:5090/",
    apiVersion: "v4",
    apiRole: "client",
    authMethod: AUTH_METHOD_PASSWORD,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      serverUrl: z.url({ protocol: z.core.regexes.httpProtocol }),
      apiVersion: z.enum(["v3", "v4"]),
      apiRole: z.enum(["client", "master"]),
      authMethod: z.enum([AUTH_METHOD_PASSWORD, AUTH_METHOD_APIKEY]),
      username: z.string().nullish(),
      password: z.string().nullish(),
      apiKey: z.string().nullish(),
      allowInsecureConnections: z.boolean().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.authMethod) {
        case AUTH_METHOD_PASSWORD:
          {
            const scUsername = z.string().nonempty();
            const spUsername = scUsername.safeParse(values.username);
            if (!spUsername.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spUsername.error).errors.join(),
                path: ["username"],
              });
            }

            const scPassword = z.string().nonempty();
            const spPassword = scPassword.safeParse(values.password);
            if (!spPassword.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPassword.error).errors.join(),
                path: ["password"],
              });
            }
          }
          break;

        case AUTH_METHOD_APIKEY:
          {
            const scApiKey = z.string().nonempty();
            const spApiKey = scApiKey.safeParse(values.apiKey);
            if (!spApiKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spApiKey.error).errors.join(),
                path: ["apiKey"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderLeCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
