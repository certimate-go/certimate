import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormVolcEngineImageXConfigFieldValues = Nullish<{
  domain: string;
}>;

export type DeployNodeConfigFormVolcEngineImageXConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormVolcEngineImageXConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormVolcEngineImageXConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormVolcEngineImageXConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormVolcEngineImageXConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormVolcEngineImageXConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.volcengine_imagex_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.volcengine_imagex_region.placeholder")),
    serviceId: z
      .string(t("workflow_node.deploy.form.volcengine_imagex_service_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.volcengine_imagex_service_id.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.volcengine_imagex_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.volcengine_imagex_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_imagex_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_imagex_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="serviceId"
        label={t("workflow_node.deploy.form.volcengine_imagex_service_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_imagex_service_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_imagex_service_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.volcengine_imagex_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_imagex_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_imagex_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormVolcEngineImageXConfig;
