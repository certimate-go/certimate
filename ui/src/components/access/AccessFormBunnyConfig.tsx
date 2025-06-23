import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForBunny } from "@/domain/access";

type AccessFormBunnyConfigFieldValues = Nullish<AccessConfigForBunny>;

export type AccessFormBunnyConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormBunnyConfigFieldValues;
  onValuesChange?: (values: AccessFormBunnyConfigFieldValues) => void;
};

const initFormModel = (): AccessFormBunnyConfigFieldValues => {
  return {
    apiKey: "",
  };
};

const AccessFormBunnyConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormBunnyConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z.string().nonempty(t("access.form.bunny_api_key.placeholder")),
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
        label={t("access.form.bunny_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.bunny_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.bunny_api_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormBunnyConfig;
