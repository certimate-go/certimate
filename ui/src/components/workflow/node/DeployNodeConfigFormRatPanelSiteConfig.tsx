import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormRatPanelSiteConfigFieldValues = Nullish<{
  siteName: string;
}>;

export type DeployNodeConfigFormRatPanelSiteConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormRatPanelSiteConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormRatPanelSiteConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormRatPanelSiteConfigFieldValues => {
  return {
    siteName: "",
  };
};

const DeployNodeConfigFormRatPanelSiteConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormRatPanelSiteConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    siteName: z.string().nonempty(t("workflow_node.deploy.form.ratpanel_site_name.placeholder")),
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
        name="siteName"
        label={t("workflow_node.deploy.form.ratpanel_site_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ratpanel_site_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ratpanel_site_name.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormRatPanelSiteConfig;
