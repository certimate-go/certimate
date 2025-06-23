import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForNS1 } from "@/domain/access";

type AccessFormNS1ConfigFieldValues = Nullish<AccessConfigForNS1>;

export type AccessFormNS1ConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormNS1ConfigFieldValues;
  onValuesChange?: (values: AccessFormNS1ConfigFieldValues) => void;
};

const initFormModel = (): AccessFormNS1ConfigFieldValues => {
  return {
    apiKey: "",
  };
};

const AccessFormNS1Config = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormNS1ConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z
      .string()
      .min(1, t("access.form.ns1_api_key.placeholder"))
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
        name="apiKey"
        label={t("access.form.ns1_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ns1_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ns1_api_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormNS1Config;
