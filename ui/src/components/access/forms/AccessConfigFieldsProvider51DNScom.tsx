import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProvider51DNScom = () => {
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
        name={[parentNamePath, "apiKey"]}
        initialValue={initialValues.apiKey}
        label={t("access.form.51dnscom_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.51dnscom_api_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.51dnscom_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiSecret"]}
        initialValue={initialValues.apiSecret}
        label={t("access.form.51dnscom_api_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.51dnscom_api_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.51dnscom_api_secret.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    apiKey: "",
    apiSecret: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    apiKey: z.string().nonempty(t("access.form.51dnscom_api_key.placeholder")),
    apiSecret: z.string().nonempty(t("access.form.51dnscom_api_secret.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProvider51DNScom, {
  getInitialValues,
  getSchema,
});

export default _default;
