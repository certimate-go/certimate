import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForCloudflare } from "@/domain/access";

type AccessFormCloudflareConfigFieldValues = Nullish<AccessConfigForCloudflare>;

export type AccessFormCloudflareConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormCloudflareConfigFieldValues;
  onValuesChange?: (values: AccessFormCloudflareConfigFieldValues) => void;
};

const initFormModel = (): AccessFormCloudflareConfigFieldValues => {
  return {
    dnsApiToken: "",
  };
};

const AccessFormCloudflareConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormCloudflareConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    dnsApiToken: z
      .string()
      .min(1, t("access.form.cloudflare_dns_api_token.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    zoneApiToken: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish(),
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
        name="dnsApiToken"
        label={t("access.form.cloudflare_dns_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudflare_dns_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cloudflare_dns_api_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name="zoneApiToken"
        label={t("access.form.cloudflare_zone_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudflare_zone_api_token.tooltip") }}></span>}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.cloudflare_zone_api_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormCloudflareConfig;
