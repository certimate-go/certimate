import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForNamecheap } from "@/domain/access";

type AccessFormNamecheapConfigFieldValues = Nullish<AccessConfigForNamecheap>;

export type AccessFormNamecheapConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormNamecheapConfigFieldValues;
  onValuesChange?: (values: AccessFormNamecheapConfigFieldValues) => void;
};

const initFormModel = (): AccessFormNamecheapConfigFieldValues => {
  return {
    username: "",
    apiKey: "",
  };
};

const AccessFormNamecheapConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormNamecheapConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    username: z
      .string()
      .min(1, t("access.form.namecheap_username.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiKey: z
      .string()
      .min(1, t("access.form.namecheap_api_key.placeholder"))
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
        name="username"
        label={t("access.form.namecheap_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.namecheap_username.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.namecheap_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiKey"
        label={t("access.form.namecheap_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.namecheap_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.namecheap_api_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormNamecheapConfig;
