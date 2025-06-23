import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudSCFConfigFieldValues = Nullish<{
  endpoint?: string;
  region: string;
  domain: string;
}>;

export type DeployNodeConfigFormTencentCloudSCFConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudSCFConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudSCFConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormTencentCloudSCFConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudSCFConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudSCFConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    region: z
      .string(t("workflow_node.deploy.form.tencentcloud_scf_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_scf_region.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.tencentcloud_scf_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.tencentcloud_scf_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_scf_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_scf_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.tencentcloud_scf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_scf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_scf_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.tencentcloud_scf_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_scf_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_scf_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudSCFConfig;
