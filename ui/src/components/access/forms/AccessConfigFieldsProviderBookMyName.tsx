import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderBookMyName = () => {
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
        name={[parentNamePath, "username"]}
        initialValue={initialValues.username}
        label={t("access.form.bookmyname_username.label")}
        rules={[formRule]}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.bookmyname_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "password"]}
        initialValue={initialValues.password}
        label={t("access.form.bookmyname_password.label")}
        rules={[formRule]}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.bookmyname_password.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    username: "",
    password: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    username: z.string().nonempty(t("access.form.bookmyname_username.placeholder")),
    password: z.string().nonempty(t("access.form.bookmyname_password.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderBookMyName, {
  getInitialValues,
  getSchema,
});

export default _default;
