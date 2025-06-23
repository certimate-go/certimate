import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormVolcEngineCertCenterConfigFieldValues = Nullish<{
  region: string;
}>;

export type DeployNodeConfigFormVolcEngineCertCenterConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormVolcEngineCertCenterConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormVolcEngineCertCenterConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormVolcEngineCertCenterConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormVolcEngineCertCenterConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormVolcEngineCertCenterConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.volcengine_certcenter_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.volcengine_certcenter_region.placeholder")),
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
      <Form.Item name="region" label={t("workflow_node.deploy.form.volcengine_certcenter_region.label")} rules={[formRule]}>
        <Input placeholder={t("workflow_node.deploy.form.volcengine_certcenter_region.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormVolcEngineCertCenterConfig;
