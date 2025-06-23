import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForUniCloud } from "@/domain/access";

type AccessFormUniCloudConfigFieldValues = Nullish<AccessConfigForUniCloud>;

export type AccessFormUniCloudConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormUniCloudConfigFieldValues;
  onValuesChange?: (values: AccessFormUniCloudConfigFieldValues) => void;
};

const initFormModel = (): AccessFormUniCloudConfigFieldValues => {
  return {
    username: "",
    password: "",
  };
};

const AccessFormUniCloudConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormUniCloudConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    username: z.string().nonempty(t("access.form.unicloud_username.placeholder")),
    password: z.string().nonempty(t("access.form.unicloud_password.placeholder")),
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
        label={t("access.form.unicloud_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.unicloud_username.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.unicloud_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name="password"
        label={t("access.form.unicloud_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.unicloud_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.unicloud_password.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormUniCloudConfig;
