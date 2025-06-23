import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormBaishanCDNConfigFieldValues = Nullish<{
  domain: string;
  certificateId?: string | number;
}>;

export type DeployNodeConfigFormBaishanCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormBaishanCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormBaishanCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormBaishanCDNConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormBaishanCDNConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormBaishanCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domain: z
      .string(t("workflow_node.deploy.form.baishan_cdn_domain.placeholder"))
      .refine((v) => validDomainName(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
    certificateId: z
      .union([z.string(), z.number().int()])
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return /^\d+$/.test(v + "") && +v > 0;
      }, t("workflow_node.deploy.form.baishan_cdn_certificate_id.placeholder")),
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
        name="domain"
        label={t("workflow_node.deploy.form.baishan_cdn_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baishan_cdn_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baishan_cdn_domain.placeholder")} />
      </Form.Item>

      <Form.Item
        name="certificateId"
        label={t("workflow_node.deploy.form.baishan_cdn_certificate_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baishan_cdn_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear type="number" placeholder={t("workflow_node.deploy.form.baishan_cdn_certificate_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormBaishanCDNConfig;
