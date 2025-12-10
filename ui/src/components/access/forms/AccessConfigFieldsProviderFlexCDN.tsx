import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderFlexCDN = () => {
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
        label={t("access.form.flexcdn_server_url.label")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.flexcdn_server_url.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "apiRole"]} initialValue={initialValues.apiRole} label={t("access.form.flexcdn_api_role.label")} rules={[formRule]}>
        <Radio.Group options={["user", "admin"].map((s) => ({ label: t(`access.form.flexcdn_api_role.option.${s}.label`), value: s }))} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessKeyId"]}
        initialValue={initialValues.accessKeyId}
        label={t("access.form.flexcdn_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.flexcdn_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.flexcdn_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessKey"]}
        initialValue={initialValues.accessKey}
        label={t("access.form.flexcdn_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.flexcdn_access_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.flexcdn_access_key.placeholder")} />
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
    serverUrl: "http://<your-host-addr>:8000/",
    apiRole: "user",
    accessKeyId: "",
    accessKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    apiRole: z.literal(["user", "admin"], t("access.form.flexcdn_api_role.placeholder")),
    accessKeyId: z.string().nonempty(t("access.form.flexcdn_access_key_id.placeholder")),
    accessKey: z.string().nonempty(t("access.form.flexcdn_access_key.placeholder")),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderFlexCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
