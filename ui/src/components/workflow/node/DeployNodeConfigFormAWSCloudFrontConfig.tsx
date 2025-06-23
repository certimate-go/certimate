import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormAWSCloudFrontConfigFieldValues = Nullish<{
  region: string;
  distributionId: string;
  certificateSource: string;
}>;

export type DeployNodeConfigFormAWSCloudFrontConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAWSCloudFrontConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAWSCloudFrontConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormAWSCloudFrontConfigFieldValues => {
  return {
    certificateSource: "ACM",
  };
};

const DeployNodeConfigFormAWSCloudFrontConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormAWSCloudFrontConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string().nonempty(t("workflow_node.deploy.form.aws_cloudfront_region.placeholder")),
    distributionId: z.string().nonempty(t("workflow_node.deploy.form.aws_cloudfront_distribution_id.placeholder")),
    certificateSource: z.string().nonempty(t("workflow_node.deploy.form.aws_cloudfront_certificate_source.placeholder")),
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
        label={t("workflow_node.deploy.form.aws_cloudfront_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_cloudfront_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_cloudfront_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="distributionId"
        label={t("workflow_node.deploy.form.aws_cloudfront_distribution_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_cloudfront_distribution_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_cloudfront_distribution_id.placeholder")} />
      </Form.Item>

      <Form.Item name="certificateSource" label={t("workflow_node.deploy.form.aws_cloudfront_certificate_source.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.aws_cloudfront_certificate_source.placeholder")}>
          <Select.Option key="ACM" value="ACM">
            ACM
          </Select.Option>
          <Select.Option key="IAM" value="IAM">
            IAM
          </Select.Option>
        </Select>
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAWSCloudFrontConfig;
