import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormUpyunFileConfigFieldValues = Nullish<{
  domain: string;
}>;

export type DeployNodeConfigFormUpyunFileConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormUpyunFileConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormUpyunFileConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormUpyunFileConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormUpyunFileConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormUpyunFileConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domain: z.string(t("workflow_node.deploy.form.upyun_file_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.upyun_file.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.upyun_file_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.upyun_file_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.upyun_file_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormUpyunFileConfig;
