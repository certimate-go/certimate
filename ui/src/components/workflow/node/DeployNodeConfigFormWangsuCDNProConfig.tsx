import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormWangsuCDNProConfigFieldValues = Nullish<{
  environment: string;
  domain: string;
  certificateId?: string;
  webhookId?: string;
}>;

export type DeployNodeConfigFormWangsuCDNProConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormWangsuCDNProConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormWangsuCDNProConfigFieldValues) => void;
};

const ENVIRONMENT_PRODUCTION = "production" as const;
const ENVIRONMENT_STAGING = "stating" as const;

const initFormModel = (): DeployNodeConfigFormWangsuCDNProConfigFieldValues => {
  return {
    environment: ENVIRONMENT_PRODUCTION,
  };
};

const DeployNodeConfigFormWangsuCDNProConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormWangsuCDNProConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceType: z.literal([ENVIRONMENT_PRODUCTION, ENVIRONMENT_STAGING], t("workflow_node.deploy.form.wangsu_cdnpro_environment.placeholder")),
    domain: z
      .string(t("workflow_node.deploy.form.wangsu_cdnpro_domain.placeholder"))
      .refine((v) => validDomainName(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
    certificateId: z.string().nullish(),
    webhookId: z.string().nullish(),
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
      <Form.Item name="environment" label={t("workflow_node.deploy.form.wangsu_cdnpro_environment.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_environment.placeholder")}>
          <Select.Option key={ENVIRONMENT_PRODUCTION} value={ENVIRONMENT_PRODUCTION}>
            {t("workflow_node.deploy.form.wangsu_cdnpro_environment.option.production.label")}
          </Select.Option>
          <Select.Option key={ENVIRONMENT_STAGING} value={ENVIRONMENT_STAGING}>
            {t("workflow_node.deploy.form.wangsu_cdnpro_environment.option.staging.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.wangsu_cdnpro_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_cdnpro_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_domain.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificateId"
        label={t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="webhookId"
        label={t("workflow_node.deploy.form.wangsu_cdnpro_webhook_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_cdnpro_webhook_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_webhook_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormWangsuCDNProConfig;
