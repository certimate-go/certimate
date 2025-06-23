import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";

type DeployNodeConfigFormCdnflyConfigFieldValues = Nullish<{
  resourceType: string;
  siteId?: string | number;
  certificateId?: string | number;
}>;

export type DeployNodeConfigFormCdnflyConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormCdnflyConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormCdnflyConfigFieldValues) => void;
};

const RESOURCE_TYPE_SITE = "site" as const;
const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const initFormModel = (): DeployNodeConfigFormCdnflyConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_SITE,
  };
};

const DeployNodeConfigFormCdnflyConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormCdnflyConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceType: z.literal([RESOURCE_TYPE_SITE, RESOURCE_TYPE_CERTIFICATE], t("workflow_node.deploy.form.cdnfly_resource_type.placeholder")),
    siteId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_SITE) return true;
        return /^\d+$/.test(v + "") && +v! > 0;
      }, t("workflow_node.deploy.form.cdnfly_site_id.placeholder")),
    certificateId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_CERTIFICATE) return true;
        return /^\d+$/.test(v + "") && +v! > 0;
      }, t("workflow_node.deploy.form.cdnfly_certificate_id.placeholder")),
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
      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.cdnfly_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.cdnfly_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_SITE} value={RESOURCE_TYPE_SITE}>
            {t("workflow_node.deploy.form.cdnfly_resource_type.option.site.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_CERTIFICATE} value={RESOURCE_TYPE_CERTIFICATE}>
            {t("workflow_node.deploy.form.cdnfly_resource_type.option.certificate.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_SITE}>
        <Form.Item
          name="siteId"
          label={t("workflow_node.deploy.form.cdnfly_site_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.cdnfly_site_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.cdnfly_site_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name="certificateId"
          label={t("workflow_node.deploy.form.cdnfly_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.cdnfly_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.cdnfly_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormCdnflyConfig;
