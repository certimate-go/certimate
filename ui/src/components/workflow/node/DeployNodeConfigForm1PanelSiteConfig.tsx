import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";

type DeployNodeConfigForm1PanelSiteConfigFieldValues = Nullish<{
  resourceType: string;
  websiteId?: string | number;
  certificateId?: string | number;
}>;

export type DeployNodeConfigForm1PanelSiteConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigForm1PanelSiteConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigForm1PanelSiteConfigFieldValues) => void;
};

const RESOURCE_TYPE_WEBSITE = "website" as const;
const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const initFormModel = (): DeployNodeConfigForm1PanelSiteConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_WEBSITE,
  };
};

const DeployNodeConfigForm1PanelSiteConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigForm1PanelSiteConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceType: z.literal([RESOURCE_TYPE_WEBSITE, RESOURCE_TYPE_CERTIFICATE], t("workflow_node.deploy.form.1panel_site_resource_type.placeholder")),
    websiteId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_WEBSITE) return true;
        return /^\d+$/.test(v + "") && +v! > 0;
      }, t("workflow_node.deploy.form.1panel_site_website_id.placeholder")),
    certificateId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_CERTIFICATE) return true;
        return /^\d+$/.test(v + "") && +v! > 0;
      }, t("workflow_node.deploy.form.1panel_site_certificate_id.placeholder")),
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
      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.1panel_site_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.1panel_site_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_WEBSITE} value={RESOURCE_TYPE_WEBSITE}>
            {t("workflow_node.deploy.form.1panel_site_resource_type.option.website.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_CERTIFICATE} value={RESOURCE_TYPE_CERTIFICATE}>
            {t("workflow_node.deploy.form.1panel_site_resource_type.option.certificate.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_WEBSITE}>
        <Form.Item
          name="websiteId"
          label={t("workflow_node.deploy.form.1panel_site_website_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_site_website_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.1panel_site_website_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name="certificateId"
          label={t("workflow_node.deploy.form.1panel_site_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_site_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.1panel_site_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigForm1PanelSiteConfig;
