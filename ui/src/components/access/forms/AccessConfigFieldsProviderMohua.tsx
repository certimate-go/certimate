import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderMohua = () => {
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
        name={[parentNamePath, "accessKey"]}
        initialValue={initialValues.accessKey}
        label={t("access.form.mohua_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.mohua_access_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.mohua_access_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretKey"]}
        initialValue={initialValues.secretKey}
        label={t("access.form.mohua_secret_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.mohua_secret_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.mohua_secret_key.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    accessKey: "",
    secretKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    accessKey: z.string().nonempty(t("access.form.mohua_access_key.placeholder")),
    secretKey: z.string().nonempty(t("access.form.mohua_secret_key.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderMohua, {
  getInitialValues,
  getSchema,
});

export default _default;
