import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderClouDNS = () => {
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
        name={[parentNamePath, "authId"]}
        initialValue={initialValues.authId}
        label={t("access.form.cloudns_auth_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudns_auth_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.cloudns_auth_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "authPassword"]}
        initialValue={initialValues.authPassword}
        label={t("access.form.cloudns_auth_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudns_auth_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cloudns_auth_password.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    authId: "",
    authPassword: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    authId: z.string().nonempty(),
    authPassword: z.string().nonempty(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderClouDNS, {
  getInitialValues,
  getSchema,
});

export default _default;
