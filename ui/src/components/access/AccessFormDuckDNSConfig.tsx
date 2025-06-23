import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForDuckDNS } from "@/domain/access";

type AccessFormDuckDNSConfigFieldValues = Nullish<AccessConfigForDuckDNS>;

export type AccessFormDuckDNSConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormDuckDNSConfigFieldValues;
  onValuesChange?: (values: AccessFormDuckDNSConfigFieldValues) => void;
};

const initFormModel = (): AccessFormDuckDNSConfigFieldValues => {
  return {
    token: "",
  };
};

const AccessFormDuckDNSConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormDuckDNSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    token: z.string().nonempty(t("access.form.duckdns_token.placeholder")),
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
        name="token"
        label={t("access.form.duckdns_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.duckdns_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.duckdns_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormDuckDNSConfig;
