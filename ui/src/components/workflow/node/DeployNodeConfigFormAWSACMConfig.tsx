import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormAWSACMConfigFieldValues = Nullish<{
  region: string;
  certificateArn?: string;
}>;

export type DeployNodeConfigFormAWSACMConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAWSACMConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAWSACMConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormAWSACMConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormAWSACMConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormAWSACMConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.deploy.form.aws_acm_region.placeholder")).nonempty(t("workflow_node.deploy.form.aws_acm_region.placeholder")),
    certificateArn: z.string().nullish(),
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
        name="region"
        label={t("workflow_node.deploy.form.aws_acm_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_acm_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_acm_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificateArn"
        label={t("workflow_node.deploy.form.aws_acm_certificate_arn.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_acm_certificate_arn.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.aws_acm_certificate_arn.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAWSACMConfig;
