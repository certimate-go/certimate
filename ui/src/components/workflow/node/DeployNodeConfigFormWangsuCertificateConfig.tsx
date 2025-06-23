import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormWangsuCertificateConfigFieldValues = Nullish<{
  certificateId?: string;
}>;

export type DeployNodeConfigFormWangsuCertificateConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormWangsuCertificateConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormWangsuCertificateConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormWangsuCertificateConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormWangsuCertificateConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormWangsuCertificateConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    certificateId: z.string().nullish(),
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
        name="certificateId"
        label={t("workflow_node.deploy.form.wangsu_certificate_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.wangsu_certificate_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormWangsuCertificateConfig;
