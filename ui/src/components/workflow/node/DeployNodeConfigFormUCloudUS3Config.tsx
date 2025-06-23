import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormUCloudUS3ConfigFieldValues = Nullish<{
  region: string;
  bucket: string;
  domain: string;
}>;

export type DeployNodeConfigFormUCloudUS3ConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormUCloudUS3ConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormUCloudUS3ConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormUCloudUS3ConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormUCloudUS3Config = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormUCloudUS3ConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.deploy.form.ucloud_us3_region.placeholder")).nonempty(t("workflow_node.deploy.form.ucloud_us3_region.placeholder")),
    bucket: z.string(t("workflow_node.deploy.form.ucloud_us3_bucket.placeholder")).nonempty(t("workflow_node.deploy.form.ucloud_us3_bucket.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.ucloud_us3_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.ucloud_us3_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_us3_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_us3_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="bucket"
        label={t("workflow_node.deploy.form.ucloud_us3_bucket.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_us3_bucket.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_us3_bucket.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.ucloud_us3_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_us3_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_us3_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormUCloudUS3Config;
