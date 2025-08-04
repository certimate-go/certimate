import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";
import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";

type DeployNodeConfigFormTencentCloudEOConfigFieldValues = Nullish<{
  endpoint?: string;
  zoneId: string;
  domains: string;
  deployToAllDomains?: boolean;
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

const MULTIPLE_INPUT_SEPARATOR = ";";

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
    deployToAllDomains: z.boolean().nullish(),
    domains: z
      .string(t("workflow_node.deploy.form.tencentcloud_eo_domains.placeholder"))
      .refine((v) => {
        const deployToAllDomains = formInst.getFieldValue('deployToAllDomains');
        // 如果启用了"部署到全部网站"，则域名字段可以为空
        if (deployToAllDomains) return true;
        // 否则必须填写有效域名
        if (!v) return false;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => validDomainName(e, { allowWildcard: true }));
      }, t("common.errmsg.domain_invalid")),
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
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_eo_endpoint.placeholder")} />
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
        name="deployToAllDomains"
        label={t("workflow_node.deploy.form.tencentcloud_eo_deploy_to_all_domains.label")}
        valuePropName="checked"
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_deploy_to_all_domains.tooltip") }}></span>}
      >
        <Switch />
      </Form.Item>

      <Form.Item
        dependencies={['deployToAllDomains']}
      >
        {({ getFieldValue }) => {
          const deployToAllDomains = getFieldValue('deployToAllDomains');
          return (
            <Form.Item
              name="domains"
              label={t("workflow_node.deploy.form.tencentcloud_eo_domains.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_domains.tooltip") }}></span>}
              style={{ display: deployToAllDomains ? 'none' : 'block' }}
            >
              <MultipleSplitValueInput
                modalTitle={t("workflow_node.deploy.form.tencentcloud_eo_domains.multiple_input_modal.title")}
                placeholder={t("workflow_node.deploy.form.tencentcloud_eo_domains.placeholder")}
                placeholderInModal={t("workflow_node.deploy.form.tencentcloud_eo_domains.multiple_input_modal.placeholder")}
                splitOptions={{ trim: true, removeEmpty: true }}
              />
            </Form.Item>
          );
        }}
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudEOConfig;