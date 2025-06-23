import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForSpaceship } from "@/domain/access";

type AccessFormSpaceshipConfigFieldValues = Nullish<AccessConfigForSpaceship>;

export type AccessFormSpaceshipConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormSpaceshipConfigFieldValues;
  onValuesChange?: (values: AccessFormSpaceshipConfigFieldValues) => void;
};

const initFormModel = (): AccessFormSpaceshipConfigFieldValues => {
  return {
    apiKey: "",
    apiSecret: "",
  };
};

const AccessFormSpaceshipConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormSpaceshipConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z.string().nonempty(t("access.form.spaceship_api_key.placeholder")),
    apiSecret: z.string().nonempty(t("access.form.spaceship_api_secret.placeholder")),
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
        label={t("access.form.spaceship_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.spaceship_api_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.spaceship_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiSecret"
        label={t("access.form.spaceship_api_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.spaceship_api_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.spaceship_api_secret.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormSpaceshipConfig;
