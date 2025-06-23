import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormAWSIAMConfigFieldValues = Nullish<{
  region: string;
  certificatePath?: string;
}>;

export type DeployNodeConfigFormAWSIAMConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAWSIAMConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAWSIAMConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormAWSIAMConfigFieldValues => {
  return {
    certificatePath: "/",
  };
};

const DeployNodeConfigFormAWSIAMConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormAWSIAMConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.deploy.form.aws_iam_region.placeholder")).nonempty(t("workflow_node.deploy.form.aws_iam_region.placeholder")),
    certificatePath: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return v.startsWith("/") && v.endsWith("/");
      }, t("workflow_node.deploy.form.aws_iam_certificate_path.errmsg.invalid")),
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
        label={t("workflow_node.deploy.form.aws_iam_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_iam_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_iam_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificatePath"
        label={t("workflow_node.deploy.form.aws_iam_certificate_path.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_iam_certificate_path.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.aws_iam_certificate_path.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAWSIAMConfig;
