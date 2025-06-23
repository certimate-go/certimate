import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormEdgioApplicationsConfigFieldValues = Nullish<{
  environmentId: string;
}>;

export type DeployNodeConfigFormEdgioApplicationsConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormEdgioApplicationsConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormEdgioApplicationsConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormEdgioApplicationsConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormEdgioApplicationsConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormEdgioApplicationsConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    environmentId: z
      .string(t("workflow_node.deploy.form.edgio_applications_environment_id.placeholder"))
      .min(1, t("workflow_node.deploy.form.edgio_applications_environment_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
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
        name="environmentId"
        label={t("workflow_node.deploy.form.edgio_applications_environment_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.edgio_applications_environment_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.edgio_applications_environment_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormEdgioApplicationsConfig;
