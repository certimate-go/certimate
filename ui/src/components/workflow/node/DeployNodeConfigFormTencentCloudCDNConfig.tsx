import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudCDNConfigFieldValues = Nullish<{
  endpoint?: string;
  domain: string;
}>;

export type DeployNodeConfigFormTencentCloudCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormTencentCloudCDNConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudCDNConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    domain: z
      .string(t("workflow_node.deploy.form.tencentcloud_cdn_domain.placeholder"))
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
        label={t("workflow_node.deploy.form.tencentcloud_cdn_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_cdn_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_cdn_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.tencentcloud_cdn_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_cdn_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_cdn_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudCDNConfig;
