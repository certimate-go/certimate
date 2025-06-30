import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";

type DeployNodeConfigFormTencentCloudSSLUpdateConfigFieldValues = Nullish<{
  endpoint?: string;
  certificateId: string;
  resourceTypes: string;
  resourceRegions?: string;
  isReplaced?: boolean;
}>;

export type DeployNodeConfigFormTencentCloudSSLUpdateConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudSSLUpdateConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudSSLUpdateConfigFieldValues) => void;
};

const MULTIPLE_INPUT_SEPARATOR = ";";

const initFormModel = (): DeployNodeConfigFormTencentCloudSSLUpdateConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudSSLUpdateConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudSSLUpdateConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    certificateId: z
      .string(t("workflow_node.deploy.form.tencentcloud_ssl_update_certificate_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_ssl_update_certificate_id.placeholder")),
    resourceTypes: z.string(t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.placeholder")).refine((v) => {
      if (!v) return false;
      return String(v)
        .split(MULTIPLE_INPUT_SEPARATOR)
        .every((e) => !!e.trim());
    }, t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.placeholder")),
    resourceRegions: z
      .string(t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.placeholder"))
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => !!e.trim());
      }, t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.placeholder")),
    isReplaced: z.boolean().nullish(),
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
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_update.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name="endpoint"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_update_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_update_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_update_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificateId"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_update_certificate_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_update_certificate_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_update_certificate_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="resourceTypes"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_types.multiple_input_modal.placeholder")}
          splitOptions={{ trim: true, removeEmpty: true }}
        />
      </Form.Item>

      <Form.Item
        name="resourceRegions"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.tencentcloud_ssl_update_resource_regions.multiple_input_modal.placeholder")}
          splitOptions={{ trim: true, removeEmpty: true }}
        />
      </Form.Item>

      <Form.Item
        name="isReplaced"
        label={t("workflow_node.deploy.form.tencentcloud_ssl_update_is_replaced.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssl_update_is_replaced.tooltip") }}></span>}
      >
        <Switch />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudSSLUpdateConfig;
