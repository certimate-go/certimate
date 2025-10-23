import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderIONOS = () => {
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
        name={[parentNamePath, "apiKeyPublicPrefix"]}
        initialValue={initialValues.apiKeyPublicPrefix}
        label={t("access.form.ionos_api_key_public_prefix.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ionos_api_key_public_prefix.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.ionos_api_key_public_prefix.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiKeySecret"]}
        initialValue={initialValues.apiKeySecret}
        label={t("access.form.ionos_api_key_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ionos_api_key_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ionos_api_key_secret.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    apiKeyPublicPrefix: "",
    apiKeySecret: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    apiKeyPublicPrefix: z.string().nonempty(t("access.form.ionos_api_key_public_prefix.placeholder")),
    apiKeySecret: z.string().nonempty(t("access.form.ionos_api_key_secret.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderIONOS, {
  getInitialValues,
  getSchema,
});

export default _default;
