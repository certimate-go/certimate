import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormGcoreCDNConfigFieldValues = Nullish<{
  resourceId: string | number;
  certificateId?: string | number;
}>;

export type DeployNodeConfigFormGcoreCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormGcoreCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormGcoreCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormGcoreCDNConfigFieldValues => {
  return {
    resourceId: "",
  };
};

const DeployNodeConfigFormGcoreCDNConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormGcoreCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceId: z.union([z.string(), z.number().int()]).refine((v) => {
      return /^\d+$/.test(v + "") && +v > 0;
    }, t("workflow_node.deploy.form.gcore_cdn_resource_id.placeholder")),
    certificateId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return /^\d+$/.test(v + "") && +v > 0;
      }, t("workflow_node.deploy.form.gcore_cdn_certificate_id.placeholder")),
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
        name="resourceId"
        label={t("workflow_node.deploy.form.gcore_cdn_resource_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.gcore_cdn_resource_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("workflow_node.deploy.form.gcore_cdn_resource_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificateId"
        label={t("workflow_node.deploy.form.gcore_cdn_certificate_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.gcore_cdn_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear type="number" placeholder={t("workflow_node.deploy.form.gcore_cdn_certificate_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormGcoreCDNConfig;
