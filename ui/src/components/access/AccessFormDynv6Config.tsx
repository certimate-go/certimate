import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForDynv6 } from "@/domain/access";

type AccessFormDynv6ConfigFieldValues = Nullish<AccessConfigForDynv6>;

export type AccessFormDynv6ConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormDynv6ConfigFieldValues;
  onValuesChange?: (values: AccessFormDynv6ConfigFieldValues) => void;
};

const initFormModel = (): AccessFormDynv6ConfigFieldValues => {
  return {
    httpToken: "",
  };
};

const AccessFormDynv6Config = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormDynv6ConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    httpToken: z
      .string()
      .min(1, t("access.form.dynv6_http_token.placeholder"))
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
        name="httpToken"
        label={t("access.form.dynv6_http_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dynv6_http_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.dynv6_http_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormDynv6Config;
