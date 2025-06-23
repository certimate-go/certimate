import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudEOConfigFieldValues = Nullish<{
  endpoint?: string;
  zoneId: string;
  domain: string;
}>;

export type DeployNodeConfigFormTencentCloudEOConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudEOConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudEOConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormTencentCloudEOConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudEOConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudEOConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    zoneId: z
      .string(t("workflow_node.deploy.form.tencentcloud_eo_zone_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_eo_zone_id.placeholder")),
    domain: z
      .string(t("workflow_node.deploy.form.tencentcloud_eo_domain.placeholder"))
      .refine((v) => validDomainName(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
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
        name="endpoint"
        label={t("workflow_node.deploy.form.tencentcloud_eo_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_eo_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="zoneId"
        label={t("workflow_node.deploy.form.tencentcloud_eo_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_eo_zone_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.tencentcloud_eo_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_eo_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudEOConfig;
