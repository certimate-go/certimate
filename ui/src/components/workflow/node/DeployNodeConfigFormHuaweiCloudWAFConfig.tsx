import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormHuaweiCloudWAFConfigFieldValues = Nullish<{
  region: string;
  resourceType: string;
  certificateId?: string;
  domain?: string;
  listenerId?: string;
}>;

export type DeployNodeConfigFormHuaweiCloudWAFConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormHuaweiCloudWAFConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormHuaweiCloudWAFConfigFieldValues) => void;
};

const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;
const RESOURCE_TYPE_CLOUDSERVER = "cloudserver" as const;
const RESOURCE_TYPE_PREMIUMHOST = "premiumhost" as const;

const initFormModel = (): DeployNodeConfigFormHuaweiCloudWAFConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormHuaweiCloudWAFConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormHuaweiCloudWAFConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.huaweicloud_waf_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.huaweicloud_waf_region.placeholder")),
    resourceType: z.literal(
      [RESOURCE_TYPE_CERTIFICATE, RESOURCE_TYPE_CLOUDSERVER, RESOURCE_TYPE_PREMIUMHOST],
      t("workflow_node.deploy.form.huaweicloud_waf_resource_type.placeholder")
    ),
    certificateId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_CERTIFICATE) return true;
        return !!v?.trim();
      }, t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.placeholder")),
    domain: z
      .string()
      .nullish()
      .refine((v) => {
        if (fieldResourceType !== RESOURCE_TYPE_CLOUDSERVER && fieldResourceType !== RESOURCE_TYPE_PREMIUMHOST) return true;
        return validDomainName(v!, { allowWildcard: true });
      }, t("workflow_node.deploy.form.huaweicloud_waf_domain.placeholder")),
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
      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.huaweicloud_waf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_waf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_waf_region.placeholder")} />
      </Form.Item>

      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.huaweicloud_waf_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.huaweicloud_waf_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_CERTIFICATE} value={RESOURCE_TYPE_CERTIFICATE}>
            {t("workflow_node.deploy.form.huaweicloud_waf_resource_type.option.certificate.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_CLOUDSERVER} value={RESOURCE_TYPE_CLOUDSERVER}>
            {t("workflow_node.deploy.form.huaweicloud_waf_resource_type.option.cloudserver.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_PREMIUMHOST} value={RESOURCE_TYPE_PREMIUMHOST}>
            {t("workflow_node.deploy.form.huaweicloud_waf_resource_type.option.premiumhost.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name="certificateId"
          label={t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_CLOUDSERVER || fieldResourceType === RESOURCE_TYPE_PREMIUMHOST}>
        <Form.Item
          name="domain"
          label={t("workflow_node.deploy.form.huaweicloud_waf_domain.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_waf_domain.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_waf_domain.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormHuaweiCloudWAFConfig;
