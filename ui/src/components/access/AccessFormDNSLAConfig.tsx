import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForDNSLA } from "@/domain/access";

type AccessFormDNSLAConfigFieldValues = Nullish<AccessConfigForDNSLA>;

export type AccessFormDNSLAConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormDNSLAConfigFieldValues;
  onValuesChange?: (values: AccessFormDNSLAConfigFieldValues) => void;
};

const initFormModel = (): AccessFormDNSLAConfigFieldValues => {
  return {
    apiId: "",
    apiSecret: "",
  };
};

const AccessFormDNSLAConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange: onValuesChange }: AccessFormDNSLAConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiId: z
      .string()
      .min(1, t("access.form.dnsla_api_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiSecret: z
      .string()
      .min(1, t("access.form.dnsla_api_secret.placeholder"))
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
        name="apiId"
        label={t("access.form.dnsla_api_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dnsla_api_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.dnsla_api_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiSecret"
        label={t("access.form.dnsla_api_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dnsla_api_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.dnsla_api_secret.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormDNSLAConfig;
