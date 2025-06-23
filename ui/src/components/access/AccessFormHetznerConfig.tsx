import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForHetzner } from "@/domain/access";

type AccessFormHetznerConfigFieldValues = Nullish<AccessConfigForHetzner>;

export type AccessFormHetznerConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormHetznerConfigFieldValues;
  onValuesChange?: (values: AccessFormHetznerConfigFieldValues) => void;
};

const initFormModel = (): AccessFormHetznerConfigFieldValues => {
  return {
    apiToken: "",
  };
};

const AccessFormHetznerConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormHetznerConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiToken: z.string().nonempty(t("access.form.hetzner_api_token.placeholder")),
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
        name="apiToken"
        label={t("access.form.hetzner_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.hetzner_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.hetzner_api_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormHetznerConfig;
