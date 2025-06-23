import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForDigitalOcean } from "@/domain/access";

type AccessFormDigitalOceanConfigFieldValues = Nullish<AccessConfigForDigitalOcean>;

export type AccessFormDigitalOceanConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormDigitalOceanConfigFieldValues;
  onValuesChange?: (values: AccessFormDigitalOceanConfigFieldValues) => void;
};

const initFormModel = (): AccessFormDigitalOceanConfigFieldValues => {
  return {
    accessToken: "",
  };
};

const AccessFormDigitalOceanConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormDigitalOceanConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessToken: z.string().nonempty(t("access.form.digitalocean_access_token.placeholder")),
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
        name="accessToken"
        label={t("access.form.digitalocean_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.digitalocean_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.digitalocean_access_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormDigitalOceanConfig;
