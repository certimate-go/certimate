import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudWAFConfigFieldValues = Nullish<{
  endpoint?: string;
  region: string;
  domain: string;
  domainId: string;
  instanceId: string;
}>;

export type DeployNodeConfigFormTencentCloudWAFConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudWAFConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudWAFConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormTencentCloudWAFConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudWAFConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudWAFConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    region: z
      .string(t("workflow_node.deploy.form.tencentcloud_waf_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_waf_region.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.tencentcloud_waf_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
    domainId: z
      .string(t("workflow_node.deploy.form.tencentcloud_waf_domain_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_waf_domain_id.placeholder")),
    instanceId: z
      .string(t("workflow_node.deploy.form.tencentcloud_waf_instance_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_waf_instance_id.placeholder")),
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
        label={t("workflow_node.deploy.form.tencentcloud_waf_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_waf_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_waf_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.tencentcloud_waf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_waf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_waf_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.tencentcloud_waf_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_waf_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_waf_domain.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domainId"
        label={t("workflow_node.deploy.form.tencentcloud_waf_domain_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_waf_domain_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_waf_domain_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="instanceId"
        label={t("workflow_node.deploy.form.tencentcloud_waf_instance_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_waf_instance_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_waf_instance_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudWAFConfig;
