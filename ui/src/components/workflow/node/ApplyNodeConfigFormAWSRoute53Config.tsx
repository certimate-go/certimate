import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type ApplyNodeConfigFormAWSRoute53ConfigFieldValues = Nullish<{
  region: string;
  hostedZoneId: string;
}>;

export type ApplyNodeConfigFormAWSRoute53ConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: ApplyNodeConfigFormAWSRoute53ConfigFieldValues;
  onValuesChange?: (values: ApplyNodeConfigFormAWSRoute53ConfigFieldValues) => void;
};

const initFormModel = (): ApplyNodeConfigFormAWSRoute53ConfigFieldValues => {
  return {
    region: "us-east-1",
    hostedZoneId: "",
  };
};

const ApplyNodeConfigFormAWSRoute53Config = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: ApplyNodeConfigFormAWSRoute53ConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.apply.form.aws_route53_region.placeholder")).nonempty(t("workflow_node.apply.form.aws_route53_region.placeholder")),
    hostedZoneId: z
      .string(t("workflow_node.apply.form.aws_route53_hosted_zone_id.placeholder"))
      .nonempty(t("workflow_node.apply.form.aws_route53_hosted_zone_id.placeholder")),
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
        label={t("workflow_node.apply.form.aws_route53_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.aws_route53_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.aws_route53_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="hostedZoneId"
        label={t("workflow_node.apply.form.aws_route53_hosted_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.aws_route53_hosted_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.aws_route53_hosted_zone_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default ApplyNodeConfigFormAWSRoute53Config;
