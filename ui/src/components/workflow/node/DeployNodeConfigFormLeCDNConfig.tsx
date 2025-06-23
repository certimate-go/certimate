import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";

type DeployNodeConfigFormLeCDNConfigFieldValues = Nullish<{
  resourceType: string;
  certificateId?: string | number;
  clientId?: string | number;
}>;

export type DeployNodeConfigFormLeCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormLeCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormLeCDNConfigFieldValues) => void;
};

const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const initFormModel = (): DeployNodeConfigFormLeCDNConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_CERTIFICATE,
    certificateId: "",
    clientId: "",
  };
};

const DeployNodeConfigFormLeCDNConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormLeCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceType: z.literal(RESOURCE_TYPE_CERTIFICATE, t("workflow_node.deploy.form.lecdn_resource_type.placeholder")),
    certificateId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_CERTIFICATE) return true;
        return /^\d+$/.test(v + "") && +v! > 0;
      }, t("workflow_node.deploy.form.lecdn_certificate_id.placeholder")),
    clientId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_CERTIFICATE) return true;
        if (v == null || v === "") return true;
        return /^\d+$/.test(v + "") && +v! > 0;
      }, t("workflow_node.deploy.form.lecdn_client_id.placeholder")),
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
      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.lecdn_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.lecdn_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_CERTIFICATE} value={RESOURCE_TYPE_CERTIFICATE}>
            {t("workflow_node.deploy.form.lecdn_resource_type.option.certificate.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name="certificateId"
          label={t("workflow_node.deploy.form.lecdn_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.lecdn_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.lecdn_certificate_id.placeholder")} />
        </Form.Item>

        <Form.Item
          name="clientId"
          label={t("workflow_node.deploy.form.lecdn_client_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.lecdn_client_id.tooltip") }}></span>}
        >
          <Input type="number" allowClear placeholder={t("workflow_node.deploy.form.lecdn_client_id.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormLeCDNConfig;
