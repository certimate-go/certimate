import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormJDCloudLiveConfigFieldValues = Nullish<{
  domain: string;
}>;

export type DeployNodeConfigFormJDCloudLiveConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormJDCloudLiveConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormJDCloudLiveConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormJDCloudLiveConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormJDCloudLiveConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormJDCloudLiveConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domain: z.string(t("workflow_node.deploy.form.jdcloud_live_domain.placeholder")).refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        label={t("workflow_node.deploy.form.jdcloud_live_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_live_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.jdcloud_live_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormJDCloudLiveConfig;
