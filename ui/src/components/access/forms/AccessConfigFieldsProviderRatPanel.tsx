import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderRatPanel = () => {
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
        label={t("access.form.ratpanel_server_url.label")}
        extra={t("access.form.ratpanel_server_url.help")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.ratpanel_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessTokenId"]}
        initialValue={initialValues.accessTokenId}
        label={t("access.form.ratpanel_access_token_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ratpanel_access_token_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("access.form.ratpanel_access_token_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessToken"]}
        initialValue={initialValues.accessToken}
        label={t("access.form.ratpanel_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ratpanel_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ratpanel_access_token.placeholder")} />
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
    serverUrl: "http://<your-host-addr>:8888/",
    accessTokenId: 1,
    accessToken: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    accessTokenId: z.coerce.number().int(t("access.form.ratpanel_access_token_id.placeholder")).positive(t("access.form.ratpanel_access_token_id.placeholder")),
    accessToken: z.string().nonempty(t("access.form.ratpanel_access_token.placeholder")),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderRatPanel, {
  getInitialValues,
  getSchema,
});

export default _default;
