import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormBaotaPanelConsoleConfigFieldValues = Nullish<{
  autoRestart?: boolean;
}>;

export type DeployNodeConfigFormBaotaPanelConsoleConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormBaotaPanelConsoleConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormBaotaPanelConsoleConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormBaotaPanelConsoleConfigFieldValues => {
  return {
    autoRestart: true,
  };
};

const DeployNodeConfigFormBaotaPanelConsoleConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormBaotaPanelConsoleConfigProps) => {
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
      <Form.Item name="autoRestart" label={t("workflow_node.deploy.form.baotapanel_console_auto_restart.label")} rules={[formRule]}>
        <Switch />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormBaotaPanelConsoleConfig;
