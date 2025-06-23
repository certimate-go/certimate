import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormQiniuPiliConfigFieldValues = Nullish<{
  hub: string;
  domain: string;
}>;

export type DeployNodeConfigFormQiniuPiliConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormQiniuPiliConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormQiniuPiliConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormQiniuPiliConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormQiniuPiliConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormQiniuPiliConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    hub: z.string(t("workflow_node.deploy.form.qiniu_pili_hub.placeholder")).nonempty(t("workflow_node.deploy.form.qiniu_pili_hub.placeholder")),
    domain: z
      .string(t("workflow_node.deploy.form.qiniu_pili_domain.placeholder"))
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
      <Form.Item
        name="hub"
        label={t("workflow_node.deploy.form.qiniu_pili_hub.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.qiniu_pili_hub.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.qiniu_pili_hub.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.qiniu_pili_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.qiniu_pili_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.qiniu_pili_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormQiniuPiliConfig;
