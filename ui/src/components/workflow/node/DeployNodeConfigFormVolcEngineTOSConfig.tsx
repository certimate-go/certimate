import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormVolcEngineTOSConfigFieldValues = Nullish<{
  region: string;
  bucket: string;
  domain: string;
}>;

export type DeployNodeConfigFormVolcEngineTOSConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormVolcEngineTOSConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormVolcEngineTOSConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormVolcEngineTOSConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormVolcEngineTOSConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormVolcEngineTOSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.volcengine_tos_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.volcengine_tos_region.placeholder")),
    bucket: z
      .string(t("workflow_node.deploy.form.volcengine_tos_bucket.placeholder"))
      .nonempty(t("workflow_node.deploy.form.volcengine_tos_bucket.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.volcengine_tos_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.volcengine_tos_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_tos_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_tos_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="bucket"
        label={t("workflow_node.deploy.form.volcengine_tos_bucket.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_tos_bucket.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_tos_bucket.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.volcengine_tos_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_tos_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_tos_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormVolcEngineTOSConfig;
