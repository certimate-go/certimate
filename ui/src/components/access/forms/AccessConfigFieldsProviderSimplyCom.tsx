import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderSimplyCom = () => {
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
        name={[parentNamePath, "accountNumber"]}
        initialValue={initialValues.accountNumber}
        label={t("access.form.simplycom_account_number.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.simplycom_account_number.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.simplycom_account_number.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiKey"]}
        initialValue={initialValues.apiKey}
        label={t("access.form.simplycom_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.simplycom_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.simplycom_api_key.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    accountNumber: "",
    apiKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    accountNumber: z.string().nonempty(),
    apiKey: z.string().nonempty(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderSimplyCom, {
  getInitialValues,
  getSchema,
});

export default _default;
