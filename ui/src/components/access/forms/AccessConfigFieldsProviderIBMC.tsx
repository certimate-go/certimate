import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFieldsProviderIBMC = () => {
  const { i18n, t } = useTranslation();
  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({ [parentNamePath]: getSchema({ i18n }) });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item name={[parentNamePath, "endpoint"]} initialValue={initialValues.endpoint} label={t("access.form.ibmc_endpoint.label")} rules={[formRule]}>
        <Input.TextArea autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("access.form.ibmc_endpoint.placeholder")} />
      </Form.Item>
      <Form.Item name={[parentNamePath, "username"]} initialValue={initialValues.username} label={t("access.form.ibmc_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.ibmc_username.placeholder")} />
      </Form.Item>
      <Form.Item name={[parentNamePath, "password"]} initialValue={initialValues.password} label={t("access.form.ibmc_password.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ibmc_password.placeholder")} />
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

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => ({
  endpoint: "<your-ibmc-host>\n",
  username: "",
  password: "",
  allowInsecureConnections: true,
});

const getSchema = ({ i18n: _i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) =>
  z.object({
    endpoint: z.string().refine((value) => value.split(/\r?\n/).some((line) => line.trim().length > 0), "Enter at least one iBMC host"),
    username: z.string().nonempty(),
    password: z.string().nonempty(),
    allowInsecureConnections: z.boolean().nullish(),
  });

export default Object.assign(AccessConfigFieldsProviderIBMC, { getInitialValues, getSchema });
