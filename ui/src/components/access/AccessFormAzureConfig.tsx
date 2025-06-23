import { useTranslation } from "react-i18next";
import { AutoComplete, Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForAzure } from "@/domain/access";

type AccessFormAzureConfigFieldValues = Nullish<AccessConfigForAzure>;

export type AccessFormAzureConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormAzureConfigFieldValues;
  onValuesChange?: (values: AccessFormAzureConfigFieldValues) => void;
};

const initFormModel = (): AccessFormAzureConfigFieldValues => {
  return {
    tenantId: "",
    clientId: "",
    clientSecret: "",
  };
};

const AccessFormAzureConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormAzureConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    tenantId: z
      .string()
      .min(1, t("access.form.azure_tenant_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    clientId: z
      .string()
      .min(1, t("access.form.azure_client_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    clientSecret: z
      .string()
      .min(1, t("access.form.azure_client_secret.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    cloudName: z.string().nullish(),
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
        name="tenantId"
        label={t("access.form.azure_tenant_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.azure_tenant_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.azure_tenant_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="clientId"
        label={t("access.form.azure_client_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.azure_client_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.azure_client_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="clientSecret"
        label={t("access.form.azure_client_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.azure_client_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.azure_client_secret.placeholder")} />
      </Form.Item>

      <Form.Item
        name="cloudName"
        label={t("access.form.azure_cloud_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.azure_cloud_name.tooltip") }}></span>}
      >
        <AutoComplete
          options={["public", "azureusgovernment", "azurechina"].map((value) => ({ value }))}
          placeholder={t("access.form.azure_cloud_name.placeholder")}
          filterOption={(inputValue, option) => option!.value.toLowerCase().includes(inputValue.toLowerCase())}
        />
      </Form.Item>
    </Form>
  );
};

export default AccessFormAzureConfig;
