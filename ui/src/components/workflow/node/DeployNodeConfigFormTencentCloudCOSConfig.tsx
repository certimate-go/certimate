import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudCOSConfigFieldValues = Nullish<{
  region: string;
  bucket: string;
  domain: string;
}>;

export type DeployNodeConfigFormTencentCloudCOSConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudCOSConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudCOSConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormTencentCloudCOSConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudCOSConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudCOSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.tencentcloud_cos_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_cos_region.placeholder")),
    bucket: z
      .string(t("workflow_node.deploy.form.tencentcloud_cos_bucket.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_cos_bucket.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.tencentcloud_cos_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.tencentcloud_cos_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_cos_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_cos_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="bucket"
        label={t("workflow_node.deploy.form.tencentcloud_cos_bucket.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_cos_bucket.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_cos_bucket.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.tencentcloud_cos_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_cos_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_cos_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudCOSConfig;
