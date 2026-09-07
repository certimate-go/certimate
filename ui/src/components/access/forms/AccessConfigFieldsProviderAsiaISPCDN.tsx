import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderAsiaISPCDN = () => {
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
        name={[parentNamePath, "accessKeyId"]}
        initialValue={initialValues.accessKeyId}
        label={t("access.form.asiaispcdn_access_key_id.label")}
        rules={[formRule]}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.asiaispcdn_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessKeySecret"]}
        initialValue={initialValues.accessKeySecret}
        label={t("access.form.asiaispcdn_access_key_secret.label")}
        rules={[formRule]}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.asiaispcdn_access_key_secret.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    accessKeyId: "",
    accessKeySecret: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    accessKeyId: z.string().nonempty(),
    accessKeySecret: z.string().nonempty(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderAsiaISPCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
