import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormQiniuKodoConfigFieldValues = Nullish<{
  domain: string;
}>;

export type DeployNodeConfigFormQiniuKodoConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormQiniuKodoConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormQiniuKodoConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormQiniuKodoConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormQiniuKodoConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormQiniuKodoConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domain: z.string(t("workflow_node.deploy.form.qiniu_kodo_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.qiniu_kodo_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.qiniu_kodo_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.qiniu_kodo_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormQiniuKodoConfig;
