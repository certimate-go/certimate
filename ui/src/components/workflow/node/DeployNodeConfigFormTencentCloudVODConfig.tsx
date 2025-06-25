import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudVODConfigFieldValues = Nullish<{
  endpoint?: string;
  subAppId?: string | number;
  domain: string;
}>;

export type DeployNodeConfigFormTencentCloudVODConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudVODConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudVODConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormTencentCloudVODConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormTencentCloudVODConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudVODConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    subAppId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (v == null) return true;
        return /^\d+$/.test(v + "") && +v > 0;
      }, t("workflow_node.deploy.form.tencentcloud_vod_sub_app_id.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.tencentcloud_vod_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.tencentcloud_vod_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_vod_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_vod_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="subAppId"
        label={t("workflow_node.deploy.form.tencentcloud_vod_sub_app_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_vod_sub_app_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("workflow_node.deploy.form.tencentcloud_vod_sub_app_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.tencentcloud_vod_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_vod_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_vod_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudVODConfig;
