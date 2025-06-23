import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForNameDotCom } from "@/domain/access";

type AccessFormNameDotComConfigFieldValues = Nullish<AccessConfigForNameDotCom>;

export type AccessFormNameDotComConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormNameDotComConfigFieldValues;
  onValuesChange?: (values: AccessFormNameDotComConfigFieldValues) => void;
};

const initFormModel = (): AccessFormNameDotComConfigFieldValues => {
  return {
    username: "",
    apiToken: "",
  };
};

const AccessFormNameDotComConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormNameDotComConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    username: z
      .string()
      .min(1, t("access.form.namedotcom_username.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiToken: z
      .string()
      .min(1, t("access.form.namedotcom_api_token.placeholder"))
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
        label={t("access.form.namedotcom_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.namedotcom_username.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.namedotcom_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiToken"
        label={t("access.form.namedotcom_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.namedotcom_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.namedotcom_api_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormNameDotComConfig;
