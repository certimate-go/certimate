import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForNetcup } from "@/domain/access";

type AccessFormNetcupConfigFieldValues = Nullish<AccessConfigForNetcup>;

export type AccessFormNetcupConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormNetcupConfigFieldValues;
  onValuesChange?: (values: AccessFormNetcupConfigFieldValues) => void;
};

const initFormModel = (): AccessFormNetcupConfigFieldValues => {
  return {
    customerNumber: "",
    apiKey: "",
    apiPassword: "",
  };
};

const AccessFormNetcupConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormNetcupConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    customerNumber: z.string().nonempty(t("access.form.netcup_customer_number.placeholder")),
    apiKey: z.string().nonempty(t("access.form.netcup_api_key.placeholder")),
    apiPassword: z.string().nonempty(t("access.form.netcup_api_password.placeholder")),
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
        name="customerNumber"
        label={t("access.form.netcup_customer_number.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.netcup_customer_number.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.netcup_customer_number.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiKey"
        label={t("access.form.netcup_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.netcup_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.netcup_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiPassword"
        label={t("access.form.netcup_api_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.netcup_api_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.netcup_api_password.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormNetcupConfig;
