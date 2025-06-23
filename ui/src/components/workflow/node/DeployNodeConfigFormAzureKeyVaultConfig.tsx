import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormAzureKeyVaultConfigFieldValues = Nullish<{
  keyvaultName: string;
  certificateName?: string;
}>;

export type DeployNodeConfigFormAzureKeyVaultConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAzureKeyVaultConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAzureKeyVaultConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormAzureKeyVaultConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormAzureKeyVaultConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormAzureKeyVaultConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    keyvaultName: z
      .string(t("workflow_node.deploy.form.azure_keyvault_name.placeholder"))
      .nonempty(t("workflow_node.deploy.form.azure_keyvault_name.placeholder")),
    certificateName: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return /^[a-zA-Z0-9-]{1,127}$/.test(v);
      }, t("workflow_node.deploy.form.azure_keyvault_certificate_name.errmsg.invalid")),
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
        name="keyvaultName"
        label={t("workflow_node.deploy.form.azure_keyvault_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.azure_keyvault_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.azure_keyvault_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificateName"
        label={t("workflow_node.deploy.form.azure_keyvault_certificate_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.azure_keyvault_certificate_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.azure_keyvault_certificate_name.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAzureKeyVaultConfig;
