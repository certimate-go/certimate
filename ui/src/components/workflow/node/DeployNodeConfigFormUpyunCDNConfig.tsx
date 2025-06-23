import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormUpyunCDNConfigFieldValues = Nullish<{
  domain: string;
}>;

export type DeployNodeConfigFormUpyunCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormUpyunCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormUpyunCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormUpyunCDNConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormUpyunCDNConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormUpyunCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domain: z
      .string(t("workflow_node.deploy.form.upyun_cdn_domain.placeholder"))
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
      <Form.Item>
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.upyun_cdn.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.upyun_cdn_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.upyun_cdn_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.upyun_cdn_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormUpyunCDNConfig;
