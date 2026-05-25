import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { core, z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderLeCDN = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

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
        <Select options={["v3"].map((s) => ({ label: s, value: s }))} placeholder={t("access.form.lecdn_api_version.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "apiRole"]} initialValue={initialValues.apiRole} label={t("access.form.lecdn_api_role.label")} rules={[formRule]}>
        <Radio.Group options={["client", "master"].map((s) => ({ label: t(`access.form.lecdn_api_role.option.${s}.label`), value: s }))} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "username"]} initialValue={initialValues.username} label={t("access.form.lecdn_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.lecdn_username.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "password"]} initialValue={initialValues.password} label={t("access.form.lecdn_password.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.lecdn_password.placeholder")} />
      </Form.Item>

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
    apiVersion: "v3",
    apiRole: "client",
    username: "",
    password: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    serverUrl: z.url({ protocol: core.regexes.httpProtocol }),
    apiVersion: z.enum(["v3"]),
    apiRole: z.enum(["client", "master"]),
    username: z.string().nonempty(),
    password: z.string().nonempty(),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderLeCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
