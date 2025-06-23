import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForWestcn } from "@/domain/access";

type AccessFormWestcnConfigFieldValues = Nullish<AccessConfigForWestcn>;

export type AccessFormWestcnConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormWestcnConfigFieldValues;
  onValuesChange?: (values: AccessFormWestcnConfigFieldValues) => void;
};

const initFormModel = (): AccessFormWestcnConfigFieldValues => {
  return {
    username: "",
    apiPassword: "",
  };
};

const AccessFormWestcnConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormWestcnConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    username: z
      .string()
      .min(1, t("access.form.westcn_username.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiPassword: z
      .string()
      .min(1, t("access.form.westcn_api_password.placeholder"))
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
        label={t("access.form.westcn_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.westcn_username.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.westcn_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiPassword"
        label={t("access.form.westcn_api_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.westcn_api_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.westcn_api_password.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormWestcnConfig;
