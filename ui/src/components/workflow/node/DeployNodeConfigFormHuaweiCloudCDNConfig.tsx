import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormHuaweiCloudCDNConfigFieldValues = Nullish<{
  region: string;
  domain: string;
}>;

export type DeployNodeConfigFormHuaweiCloudCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormHuaweiCloudCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormHuaweiCloudCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormHuaweiCloudCDNConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormHuaweiCloudCDNConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormHuaweiCloudCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.huaweicloud_cdn_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.huaweicloud_cdn_region.placeholder")),
    domain: z.string(t("workflow_node.deploy.form.huaweicloud_cdn_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.huaweicloud_cdn_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_cdn_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_cdn_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.huaweicloud_cdn_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_cdn_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_cdn_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormHuaweiCloudCDNConfig;
