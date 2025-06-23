import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormUCloudUCDNConfigFieldValues = Nullish<{
  domainId: string;
}>;

export type DeployNodeConfigFormUCloudUCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormUCloudUCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormUCloudUCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormUCloudUCDNConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormUCloudUCDNConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormUCloudUCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domainId: z
      .string(t("workflow_node.deploy.form.ucloud_ucdn_domain_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.ucloud_ucdn_domain_id.placeholder")),
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
        name="domainId"
        label={t("workflow_node.deploy.form.ucloud_ucdn_domain_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_ucdn_domain_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_ucdn_domain_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormUCloudUCDNConfig;
