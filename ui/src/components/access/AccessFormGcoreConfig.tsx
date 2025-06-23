import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForGcore } from "@/domain/access";

type AccessFormGcoreConfigFieldValues = Nullish<AccessConfigForGcore>;

export type AccessFormGcoreConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormGcoreConfigFieldValues;
  onValuesChange?: (values: AccessFormGcoreConfigFieldValues) => void;
};

const initFormModel = (): AccessFormGcoreConfigFieldValues => {
  return {
    apiToken: "",
  };
};

const AccessFormGcoreConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormGcoreConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiToken: z
      .string()
      .min(1, t("access.form.gcore_api_token.placeholder"))
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
        name="apiToken"
        label={t("access.form.gcore_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.gcore_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.gcore_api_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormGcoreConfig;
