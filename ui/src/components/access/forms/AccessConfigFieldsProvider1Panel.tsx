import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProvider1Panel = () => {
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
        label={t("access.form.1panel_server_url.label")}
        extra={t("access.form.1panel_server_url.help")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.1panel_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiVersion"]}
        initialValue={initialValues.apiVersion}
        label={t("access.form.1panel_api_version.label")}
        rules={[formRule]}
      >
        <Select options={["v1", "v2"].map((s) => ({ label: s, value: s }))} placeholder={t("access.form.1panel_api_version.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiKey"]}
        initialValue={initialValues.apiKey}
        label={t("access.form.1panel_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.1panel_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.1panel_api_key.placeholder")} />
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
    serverUrl: "http://<your-host-addr>:20410/",
    apiVersion: "v1",
    apiKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    apiVersion: z.string().nonempty(t("access.form.1panel_api_version.placeholder")),
    apiKey: z.string().nonempty(t("access.form.1panel_api_key.placeholder")),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProvider1Panel, {
  getInitialValues,
  getSchema,
});

export default _default;
