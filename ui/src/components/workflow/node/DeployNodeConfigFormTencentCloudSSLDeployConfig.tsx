import { useTranslation } from "react-i18next";
import { Alert, AutoComplete, Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";

type DeployNodeConfigFormTencentCloudSSLDeployConfigFieldValues = Nullish<{
  endpoint?: string;
  region: string;
  resourceType: string;
  resourceIds: string;
}>;

export type DeployNodeConfigFormTencentCloudSSLDeployConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudSSLDeployConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudSSLDeployConfigFieldValues) => void;
};

const MULTIPLE_INPUT_SEPARATOR = ";";

const initFormModel = (): DeployNodeConfigFormTencentCloudSSLDeployConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudSSLDeployConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudSSLDeployConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    region: z
      .string(t("workflow_node.deploy.form.tencentcloud_ssl_deploy_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_ssl_deploy_region.placeholder")),
    resourceType: z
      .string(t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_type.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_type.placeholder")),
    resourceIds: z.string(t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.placeholder")).refine((v) => {
      if (!v) return false;
      return String(v)
        .split(MULTIPLE_INPUT_SEPARATOR)
        .every((e) => /^[A-Za-z0-9*._\-|]+$/.test(e));
    }, t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.errmsg.invalid")),
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
      <Form.Item>
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_deploy.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name="endpoint"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_deploy_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_deploy_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="resourceType"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_type.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_type.tooltip") }}></span>}
      >
        <AutoComplete
          options={["apigateway", "cdn", "clb", "cos", "ddos", "lighthouse", "live", "tcb", "teo", "tke", "tse", "vod", "waf"].map((value) => ({ value }))}
          placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_type.placeholder")}
          filterOption={(inputValue, option) => option!.value.toLowerCase().includes(inputValue.toLowerCase())}
        />
      </Form.Item>

      <Form.Item
        name="resourceIds"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.tencentcloud_ssl_deploy_resource_ids.multiple_input_modal.placeholder")}
          splitOptions={{ trim: true, removeEmpty: true }}
        />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudSSLDeployConfig;
