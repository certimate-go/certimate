import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_METHOD_PASSWORD = "password" as const;
const AUTH_METHOD_TOKEN = "token" as const;

const AccessConfigFormFieldsProviderNginxProxyManager = () => {
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
      <Form.Item
        name={[parentNamePath, "serverUrl"]}
        initialValue={initialValues.serverUrl}
        label={t("access.form.nginxproxymanager_server_url.label")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.nginxproxymanager_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "authMethod"]}
        initialValue={initialValues.authMethod}
        label={t("access.form.nginxproxymanager_auth_method.label")}
        rules={[formRule]}
      >
        <Radio.Group block>
          <Radio.Button value={AUTH_METHOD_PASSWORD}>{t("access.form.nginxproxymanager_auth_method.option.password.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_TOKEN}>{t("access.form.nginxproxymanager_auth_method.option.token.label")}</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_PASSWORD}>
        <Form.Item
          name={[parentNamePath, "username"]}
          initialValue={initialValues.username}
          label={t("access.form.nginxproxymanager_username.label")}
          rules={[formRule]}
        >
          <Input autoComplete="new-password" placeholder={t("access.form.nginxproxymanager_username.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "password"]}
          initialValue={initialValues.password}
          label={t("access.form.nginxproxymanager_password.label")}
          rules={[formRule]}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.nginxproxymanager_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldAuthMethod === AUTH_METHOD_TOKEN}>
        <Form.Item
          name={[parentNamePath, "apiToken"]}
          initialValue={initialValues.apiToken}
          label={t("access.form.nginxproxymanager_api_token.label")}
          rules={[formRule]}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.nginxproxymanager_api_token.placeholder")} />
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
    serverUrl: "http://<your-host-addr>:81/",
    authMethod: AUTH_METHOD_PASSWORD,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      serverUrl: z.url(t("common.errmsg.url_invalid")),
      authMethod: z.literal([AUTH_METHOD_PASSWORD, AUTH_METHOD_TOKEN], t("access.form.nginxproxymanager_auth_method.placeholder")),
      username: z.string().nullish(),
      password: z.string().nullish(),
      apiToken: z.string().nullish(),
      allowInsecureConnections: z.boolean().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.authMethod) {
        case AUTH_METHOD_PASSWORD:
          {
            if (!values.username?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.nginxproxymanager_username.placeholder"),
                path: ["username"],
              });
            }

            if (!values.password?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.nginxproxymanager_password.placeholder"),
                path: ["password"],
              });
            }
          }
          break;

        case AUTH_METHOD_TOKEN:
          {
            if (!values.apiToken?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.nginxproxymanager_api_token.placeholder"),
                path: ["apiToken"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderNginxProxyManager, {
  getInitialValues,
  getSchema,
});

export default _default;
