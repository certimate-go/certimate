import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForPorkbun } from "@/domain/access";

type AccessFormPorkbunConfigFieldValues = Nullish<AccessConfigForPorkbun>;

export type AccessFormPorkbunConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormPorkbunConfigFieldValues;
  onValuesChange?: (values: AccessFormPorkbunConfigFieldValues) => void;
};

const initFormModel = (): AccessFormPorkbunConfigFieldValues => {
  return {
    apiKey: "",
    secretApiKey: "",
  };
};

const AccessFormPorkbunConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormPorkbunConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z
      .string()
      .min(1, t("access.form.porkbun_api_key.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    secretApiKey: z
      .string()
      .min(1, t("access.form.porkbun_secret_api_key.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
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
        label={t("access.form.porkbun_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.porkbun_api_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.porkbun_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretApiKey"
        label={t("access.form.porkbun_secret_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.porkbun_secret_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.porkbun_secret_api_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormPorkbunConfig;
