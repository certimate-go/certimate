import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForAPISIX } from "@/domain/access";

type AccessFormAPISIXConfigFieldValues = Nullish<AccessConfigForAPISIX>;

export type AccessFormAPISIXConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormAPISIXConfigFieldValues;
  onValuesChange?: (values: AccessFormAPISIXConfigFieldValues) => void;
};

const initFormModel = (): AccessFormAPISIXConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:9180/",
    apiKey: "",
  };
};

const AccessFormAPISIXConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormAPISIXConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    apiKey: z.string().nonempty(t("access.form.apisix_api_key.placeholder")),
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
      <Form.Item name="serverUrl" label={t("access.form.apisix_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.apisix_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiKey"
        label={t("access.form.apisix_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.apisix_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.apisix_api_key.placeholder")} />
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

export default AccessFormAPISIXConfig;
