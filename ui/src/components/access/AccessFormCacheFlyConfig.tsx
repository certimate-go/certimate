import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForCacheFly } from "@/domain/access";

type AccessFormCacheFlyConfigFieldValues = Nullish<AccessConfigForCacheFly>;

export type AccessFormCacheFlyConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormCacheFlyConfigFieldValues;
  onValuesChange?: (values: AccessFormCacheFlyConfigFieldValues) => void;
};

const initFormModel = (): AccessFormCacheFlyConfigFieldValues => {
  return {
    apiToken: "",
  };
};

const AccessFormCacheFlyConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormCacheFlyConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiToken: z.string().nonempty(t("access.form.cachefly_api_token.placeholder")),
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
      <Form.Item name="apiToken" label={t("access.form.cachefly_api_token.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cachefly_api_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormCacheFlyConfig;
