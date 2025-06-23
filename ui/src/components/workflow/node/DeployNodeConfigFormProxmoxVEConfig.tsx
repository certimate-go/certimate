import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormProxmoxVEConfigFieldValues = Nullish<{
  nodeName: string;
  autoRestart?: boolean;
}>;

export type DeployNodeConfigFormProxmoxVEConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormProxmoxVEConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormProxmoxVEConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormProxmoxVEConfigFieldValues => {
  return {
    autoRestart: true,
  };
};

const DeployNodeConfigFormProxmoxVEConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormProxmoxVEConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    nodeName: z.string(t("workflow_node.deploy.form.proxmoxve_node_name.placeholder")).nonempty(t("workflow_node.deploy.form.proxmoxve_node_name.placeholder")),
    autoRestart: z.boolean().nullish(),
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
      <Form.Item name="nodeName" label={t("workflow_node.deploy.form.proxmoxve_node_name.label")} rules={[formRule]}>
        <Input placeholder={t("workflow_node.deploy.form.proxmoxve_node_name.placeholder")} />
      </Form.Item>

      <Form.Item name="autoRestart" label={t("workflow_node.deploy.form.proxmoxve_auto_restart.label")} rules={[formRule]}>
        <Switch />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormProxmoxVEConfig;
