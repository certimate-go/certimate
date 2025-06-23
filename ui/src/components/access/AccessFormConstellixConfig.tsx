import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";
import { type AccessConfigForConstellix } from "@/domain/access";

type AccessFormConstellixConfigFieldValues = Nullish<AccessConfigForConstellix>;

export type AccessFormConstellixConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormConstellixConfigFieldValues;
  onValuesChange?: (values: AccessFormConstellixConfigFieldValues) => void;
};

const initFormModel = (): AccessFormConstellixConfigFieldValues => {
  return {
    apiKey: "",
    secretKey: "",
  };
};

const AccessFormConstellixConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange: onValuesChange }: AccessFormConstellixConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z.string().nonempty(t("access.form.constellix_api_key.placeholder")),
    secretKey: z.string().nonempty(t("access.form.constellix_secret_key.placeholder")),
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
        label={t("access.form.constellix_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.constellix_api_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.constellix_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretKey"
        label={t("access.form.constellix_secret_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.constellix_secret_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.constellix_secret_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormConstellixConfig;
