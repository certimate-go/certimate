import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormAliyunWAFConfigFieldValues = Nullish<{
  region: string;
  serviceVersion: string;
  instanceId: string;
  domain?: string;
}>;

export type DeployNodeConfigFormAliyunWAFConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAliyunWAFConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAliyunWAFConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormAliyunWAFConfigFieldValues => {
  return {
    serviceVersion: "3.0",
  };
};

const DeployNodeConfigFormAliyunWAFConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormAliyunWAFConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.deploy.form.aliyun_waf_region.placeholder")).nonempty(t("workflow_node.deploy.form.aliyun_waf_region.placeholder")),
    serviceVersion: z.literal("3.0", t("workflow_node.deploy.form.aliyun_waf_service_version.placeholder")),
    instanceId: z
      .string(t("workflow_node.deploy.form.aliyun_waf_instance_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.aliyun_waf_instance_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    domain: z
      .string()
      .nullish()
      .refine((v) => {
        return !v || validDomainName(v!, { allowWildcard: true });
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
        name="region"
        label={t("workflow_node.deploy.form.aliyun_waf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_waf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_waf_region.placeholder")} />
      </Form.Item>

      <Form.Item name="serviceVersion" label={t("workflow_node.deploy.form.aliyun_waf_service_version.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.aliyun_waf_service_version.placeholder")}>
          <Select.Option key="3.0" value="3.0">
            3.0
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="instanceId"
        label={t("workflow_node.deploy.form.aliyun_waf_instance_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_waf_instance_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_waf_instance_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.aliyun_waf_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_waf_domain.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.aliyun_waf_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAliyunWAFConfig;
