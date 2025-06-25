import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";

type DeployNodeConfigFormKongConfigFieldValues = Nullish<{
  resourceType: string;
  certificateId?: string;
}>;

export type DeployNodeConfigFormKongConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormKongConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormKongConfigFieldValues) => void;
};

const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const initFormModel = (): DeployNodeConfigFormKongConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_CERTIFICATE,
    certificateId: "",
  };
};

const DeployNodeConfigFormKongConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormKongConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceType: z.literal(RESOURCE_TYPE_CERTIFICATE, t("workflow_node.deploy.form.kong_resource_type.placeholder")),
    workspace: z.string().nullish(),
    certificateId: z
      .string()
      .nullish()
      .refine((v) => fieldResourceType !== RESOURCE_TYPE_CERTIFICATE || !!v?.trim(), t("workflow_node.deploy.form.kong_certificate_id.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const fieldResourceType = Form.useWatch("resourceType", formInst);

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
      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.kong_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.kong_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_CERTIFICATE} value={RESOURCE_TYPE_CERTIFICATE}>
            {t("workflow_node.deploy.form.kong_resource_type.option.certificate.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="workspace"
        label={t("workflow_node.deploy.form.kong_workspace.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.kong_workspace.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.kong_workspace.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name="certificateId"
          label={t("workflow_node.deploy.form.kong_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.kong_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.kong_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormKongConfig;
