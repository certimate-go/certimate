import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderDNSLA = () => {
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
        name={[parentNamePath, "apiId"]}
        initialValue={initialValues.apiId}
        label={t("access.form.dnsla_api_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dnsla_api_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.dnsla_api_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiSecret"]}
        initialValue={initialValues.apiSecret}
        label={t("access.form.dnsla_api_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dnsla_api_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.dnsla_api_secret.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    apiId: "",
    apiSecret: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    apiId: z.string().nonempty(),
    apiSecret: z.string().nonempty(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderDNSLA, {
  getInitialValues,
  getSchema,
});

export default _default;
