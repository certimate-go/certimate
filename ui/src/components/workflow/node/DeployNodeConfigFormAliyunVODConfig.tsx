import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormAliyunVODConfigFieldValues = Nullish<{
  region: string;
  domain: string;
}>;

export type DeployNodeConfigFormAliyunVODConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAliyunVODConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAliyunVODConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormAliyunVODConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormAliyunVODConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormAliyunVODConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.deploy.form.aliyun_vod_region.placeholder")).nonempty(t("workflow_node.deploy.form.aliyun_vod_region.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.aliyun_vod_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.aliyun_vod_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_vod_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_vod_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.aliyun_vod_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_vod_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_vod_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormAliyunVODConfig;
