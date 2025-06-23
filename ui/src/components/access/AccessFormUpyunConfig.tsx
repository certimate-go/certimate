import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForUpyun } from "@/domain/access";

type AccessFormUpyunConfigFieldValues = Nullish<AccessConfigForUpyun>;

export type AccessFormUpyunConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormUpyunConfigFieldValues;
  onValuesChange?: (values: AccessFormUpyunConfigFieldValues) => void;
};

const initFormModel = (): AccessFormUpyunConfigFieldValues => {
  return {
    username: "",
    password: "",
  };
};

const AccessFormUpyunConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormUpyunConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    username: z
      .string()
      .min(1, t("access.form.upyun_username.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    password: z
      .string()
      .min(1, t("access.form.upyun_password.placeholder"))
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
        label={t("access.form.upyun_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.upyun_username.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.upyun_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name="password"
        label={t("access.form.upyun_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.upyun_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.upyun_password.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormUpyunConfig;
