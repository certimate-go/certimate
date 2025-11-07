import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderAkamai = () => {
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
        name={[parentNamePath, "host"]}
        initialValue={initialValues.host}
        label={t("access.form.akamai_host.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.akamai_host.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.akamai_host.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "clientToken"]}
        initialValue={initialValues.clientToken}
        label={t("access.form.akamai_client_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.akamai_client_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.akamai_client_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "clientSecret"]}
        initialValue={initialValues.clientSecret}
        label={t("access.form.akamai_client_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.akamai_client_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.akamai_client_secret.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessToken"]}
        initialValue={initialValues.accessToken}
        label={t("access.form.akamai_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.akamai_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.akamai_access_token.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    host: "",
    clientToken: "",
    clientSecret: "",
    accessToken: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    host: z.string().nonempty(t("access.form.akamai_host.placeholder")),
    clientToken: z.string().nonempty(t("access.form.akamai_client_token.placeholder")),
    clientSecret: z.string().nonempty(t("access.form.akamai_client_secret.placeholder")),
    accessToken: z.string().nonempty(t("access.form.akamai_access_token.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderAkamai, {
  getInitialValues,
  getSchema,
});

export default _default;
