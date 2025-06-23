import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForProxmoxVE } from "@/domain/access";

type AccessFormProxmoxVEConfigFieldValues = Nullish<AccessConfigForProxmoxVE>;

export type AccessFormProxmoxVEConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormProxmoxVEConfigFieldValues;
  onValuesChange?: (values: AccessFormProxmoxVEConfigFieldValues) => void;
};

const initFormModel = (): AccessFormProxmoxVEConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:8006/",
    apiToken: "",
  };
};

const AccessFormProxmoxVEConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormProxmoxVEConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    apiToken: z.string().nonempty(t("access.form.proxmoxve_api_token.placeholder")),
    apiTokenSecret: z.string().nullish(),
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
      <Form.Item name="serverUrl" label={t("access.form.proxmoxve_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.proxmoxve_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiToken"
        label={t("access.form.proxmoxve_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.proxmoxve_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.proxmoxve_api_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiTokenSecret"
        label={t("access.form.proxmoxve_api_token_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.proxmoxve_api_token_secret.tooltip") }}></span>}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.proxmoxve_api_token_secret.placeholder")} />
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

export default AccessFormProxmoxVEConfig;
