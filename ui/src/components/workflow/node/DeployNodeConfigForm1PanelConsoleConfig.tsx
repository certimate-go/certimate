import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigForm1PanelConsoleConfigFieldValues = Nullish<{
  autoRestart?: boolean;
}>;

export type DeployNodeConfigForm1PanelConsoleConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigForm1PanelConsoleConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigForm1PanelConsoleConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigForm1PanelConsoleConfigFieldValues => {
  return {
    autoRestart: true,
  };
};

const DeployNodeConfigForm1PanelConsoleConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigForm1PanelConsoleConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
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
      <Form.Item name="autoRestart" label={t("workflow_node.deploy.form.1panel_console_auto_restart.label")} rules={[formRule]}>
        <Switch />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigForm1PanelConsoleConfig;
