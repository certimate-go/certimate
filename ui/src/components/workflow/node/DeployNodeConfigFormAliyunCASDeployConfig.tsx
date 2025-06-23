import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";

type DeployNodeConfigFormAliyunCASDeployConfigFieldValues = Nullish<{
  region: string;
  resourceIds: string;
  contactIds?: string;
}>;

export type DeployNodeConfigFormAliyunCASDeployConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAliyunCASDeployConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAliyunCASDeployConfigFieldValues) => void;
};

const MULTIPLE_INPUT_SEPARATOR = ";";

const initFormModel = (): DeployNodeConfigFormAliyunCASDeployConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormAliyunCASDeployConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormAliyunCASDeployConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.aliyun_cas_deploy_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.aliyun_cas_deploy_region.placeholder")),
    resourceIds: z.string(t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.placeholder")).refine((v) => {
      if (!v) return false;
      return String(v)
        .split(MULTIPLE_INPUT_SEPARATOR)
        .every((e) => /^[1-9]\d*$/.test(e));
    }, t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.errmsg.invalid")),
    contactIds: z
      .string(t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.placeholder"))
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => /^[1-9]\d*$/.test(e));
      }, t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.errmsg.invalid")),
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
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_cas_deploy.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.aliyun_cas_deploy_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_cas_deploy_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_cas_deploy_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="resourceIds"
        label={t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.aliyun_cas_deploy_resource_ids.multiple_input_modal.placeholder")}
          splitOptions={{ trim: true, removeEmpty: true }}
        />
      </Form.Item>

      <Form.Item
        name="contactIds"
        label={t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.aliyun_cas_deploy_contact_ids.multiple_input_modal.placeholder")}
          splitOptions={{ trim: true, removeEmpty: true }}
        />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAliyunCASDeployConfig;
