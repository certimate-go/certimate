import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForBaotaWAF } from "@/domain/access";

type AccessFormBaotaWAFConfigFieldValues = Nullish<AccessConfigForBaotaWAF>;

export type AccessFormBaotaWAFConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormBaotaWAFConfigFieldValues;
  onValuesChange?: (values: AccessFormBaotaWAFConfigFieldValues) => void;
};

const initFormModel = (): AccessFormBaotaWAFConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:8379/",
    apiKey: "",
  };
};

const AccessFormBaotaWAFConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormBaotaWAFConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    apiKey: z.string().nonempty(t("access.form.baotawaf_api_key.placeholder")),
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
      <Form.Item name="serverUrl" label={t("access.form.baotawaf_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.baotawaf_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiKey"
        label={t("access.form.baotawaf_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.baotawaf_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.baotawaf_api_key.placeholder")} />
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

export default AccessFormBaotaWAFConfig;
