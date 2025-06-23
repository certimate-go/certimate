import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForGoDaddy } from "@/domain/access";

type AccessFormGoDaddyConfigFieldValues = Nullish<AccessConfigForGoDaddy>;

export type AccessFormGoDaddyConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormGoDaddyConfigFieldValues;
  onValuesChange?: (values: AccessFormGoDaddyConfigFieldValues) => void;
};

const initFormModel = (): AccessFormGoDaddyConfigFieldValues => {
  return {
    apiKey: "",
    apiSecret: "",
  };
};

const AccessFormGoDaddyConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormGoDaddyConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z
      .string()
      .min(1, t("access.form.godaddy_api_key.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiSecret: z
      .string()
      .min(1, t("access.form.godaddy_api_secret.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
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
      <Form.Item
        name="apiKey"
        label={t("access.form.godaddy_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.godaddy_api_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.godaddy_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiSecret"
        label={t("access.form.godaddy_api_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.godaddy_api_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.godaddy_api_secret.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormGoDaddyConfig;
