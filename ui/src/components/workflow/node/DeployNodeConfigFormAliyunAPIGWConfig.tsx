import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormAliyunAPIGWConfigFieldValues = Nullish<{
  serviceType: string;
  region: string;
  gatewayId?: string;
  groupId?: string;
  domain?: string;
}>;

export type DeployNodeConfigFormAliyunAPIGWConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAliyunAPIGWConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAliyunAPIGWConfigFieldValues) => void;
};

const SERVICE_TYPE_CLOUDNATIVE = "cloudnative" as const;
const SERVICE_TYPE_TRADITIONAL = "traditional" as const;

const initFormModel = (): DeployNodeConfigFormAliyunAPIGWConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormAliyunAPIGWConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormAliyunAPIGWConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serviceType: z.literal([SERVICE_TYPE_CLOUDNATIVE, SERVICE_TYPE_TRADITIONAL], t("workflow_node.deploy.form.aliyun_apigw_service_type.placeholder")),
    region: z.string(t("workflow_node.deploy.form.aliyun_apigw_region.placeholder")).nonempty(t("workflow_node.deploy.form.aliyun_apigw_region.placeholder")),
    gatewayId: z
      .string()
      .nullish()
      .refine((v) => fieldServiceType !== SERVICE_TYPE_CLOUDNATIVE || !!v?.trim(), t("workflow_node.deploy.form.aliyun_apigw_gateway_id.placeholder")),
    groupId: z
      .string()
      .nullish()
      .refine((v) => fieldServiceType !== SERVICE_TYPE_TRADITIONAL || !!v?.trim(), t("workflow_node.deploy.form.aliyun_apigw_group_id.placeholder")),
    domain: z
      .string()
      .nonempty(t("workflow_node.deploy.form.aliyun_apigw_domain.placeholder"))
      .refine((v) => validDomainName(v!, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const fieldServiceType = Form.useWatch("serviceType", formInst);

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
      <Form.Item name="serviceType" label={t("workflow_node.deploy.form.aliyun_apigw_service_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.aliyun_apigw_service_type.placeholder")}>
          <Select.Option key={SERVICE_TYPE_CLOUDNATIVE} value={SERVICE_TYPE_CLOUDNATIVE}>
            {t("workflow_node.deploy.form.aliyun_apigw_service_type.option.cloudnative.label")}
          </Select.Option>
          <Select.Option key={SERVICE_TYPE_TRADITIONAL} value={SERVICE_TYPE_TRADITIONAL}>
            {t("workflow_node.deploy.form.aliyun_apigw_service_type.option.traditional.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.aliyun_apigw_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_region.placeholder")} />
      </Form.Item>

      <Show when={fieldServiceType === SERVICE_TYPE_CLOUDNATIVE}>
        <Form.Item
          name="gatewayId"
          label={t("workflow_node.deploy.form.aliyun_apigw_gateway_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_gateway_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_gateway_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldServiceType === SERVICE_TYPE_TRADITIONAL}>
        <Form.Item
          name="groupId"
          label={t("workflow_node.deploy.form.aliyun_apigw_group_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_group_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_group_id.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.aliyun_apigw_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAliyunAPIGWConfig;
