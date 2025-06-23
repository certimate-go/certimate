import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Radio, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForLeCDN } from "@/domain/access";

type AccessFormLeCDNConfigFieldValues = Nullish<AccessConfigForLeCDN>;

export type AccessFormLeCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormLeCDNConfigFieldValues;
  onValuesChange?: (values: AccessFormLeCDNConfigFieldValues) => void;
};

const initFormModel = (): AccessFormLeCDNConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:5090/",
    apiVersion: "v3",
    apiRole: "user",
    username: "",
    password: "",
  };
};

const AccessFormLeCDNConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormLeCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    role: z.literal(["client", "master"], t("access.form.lecdn_api_role.placeholder")),
    username: z.string().nonempty(t("access.form.lecdn_username.placeholder")),
    password: z.string().nonempty(t("access.form.lecdn_password.placeholder")),
    allowInsecureConnections: z.boolean().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
    onValuesChange?.(values);
  };

  return (
    <Form
      form={formInst}
      disabled={disabled}
      initialValues={initialValues ?? initFormModel()}
      layout="vertical"
      name={formName}
      onValuesChange={handleFormChange}
    >
      <Form.Item name="serverUrl" label={t("access.form.lecdn_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.lecdn_server_url.placeholder")} />
      </Form.Item>

      <Form.Item name="apiVersion" label={t("access.form.lecdn_api_version.label")} rules={[formRule]}>
        <Select options={["v3"].map((s) => ({ label: s, value: s }))} placeholder={t("access.form.lecdn_api_version.placeholder")} />
      </Form.Item>

      <Form.Item name="apiRole" label={t("access.form.lecdn_api_role.label")} rules={[formRule]}>
        <Radio.Group options={["user", "master"].map((s) => ({ label: t(`access.form.lecdn_api_role.option.${s}.label`), value: s }))} />
      </Form.Item>

      <Form.Item name="username" label={t("access.form.lecdn_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.lecdn_username.placeholder")} />
      </Form.Item>

      <Form.Item name="password" label={t("access.form.lecdn_password.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.lecdn_password.placeholder")} />
      </Form.Item>

      <Form.Item name="allowInsecureConnections" label={t("access.form.common_allow_insecure_conns.label")} rules={[formRule]}>
        <Switch
          checkedChildren={t("access.form.common_allow_insecure_conns.switch.on")}
          unCheckedChildren={t("access.form.common_allow_insecure_conns.switch.off")}
        />
      </Form.Item>
    </Form>
  );
};

export default AccessFormLeCDNConfig;
