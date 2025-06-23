import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormNetlifySiteConfigFieldValues = Nullish<{
  siteId: string;
}>;

export type DeployNodeConfigFormNetlifySiteConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormNetlifySiteConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormNetlifySiteConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormNetlifySiteConfigFieldValues => {
  return {
    siteId: "",
  };
};

const DeployNodeConfigFormNetlifySiteConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormNetlifySiteConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    siteId: z.string().nonempty(t("workflow_node.deploy.form.netlify_site_id.placeholder")),
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
        name="siteId"
        label={t("workflow_node.deploy.form.netlify_site_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.netlify_site_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.netlify_site_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormNetlifySiteConfig;
