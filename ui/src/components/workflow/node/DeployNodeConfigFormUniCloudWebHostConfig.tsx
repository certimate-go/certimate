import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormUniCloudWebHostConfigFieldValues = Nullish<{
  spaceProvider: string;
  spaceId: string;
  domain: string;
}>;

export type DeployNodeConfigFormUniCloudWebHostConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormUniCloudWebHostConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormUniCloudWebHostConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormUniCloudWebHostConfigFieldValues => {
  return {
    spaceProvider: "tencent",
    spaceId: "",
    domain: "",
  };
};

const DeployNodeConfigFormUniCloudWebHostConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormUniCloudWebHostConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    spaceProvider: z.string().nonempty(t("workflow_node.deploy.form.unicloud_webhost_space_provider.placeholder")),
    spaceId: z.string().nonempty(t("workflow_node.deploy.form.unicloud_webhost_space_id.placeholder")),
    domain: z.string().refine((v) => validDomainName(v), t("common.errmsg.domain_invalid")),
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
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.unicloud_webhost.guide") }}></span>} />
      </Form.Item>

      <Form.Item name="spaceProvider" label={t("workflow_node.deploy.form.unicloud_webhost_space_provider.label")} rules={[formRule]}>
        <Select
          options={["aliyun", "tencent"].map((s) => ({
            label: t(`workflow_node.deploy.form.unicloud_webhost_space_provider.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.unicloud_webhost_space_provider.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name="spaceId"
        label={t("workflow_node.deploy.form.unicloud_webhost_space_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.unicloud_webhost_space_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.unicloud_webhost_space_id.placeholder")} />
      </Form.Item>

      <Form.Item name="domain" label={t("workflow_node.deploy.form.unicloud_webhost_domain.label")} rules={[formRule]}>
        <Input placeholder={t("workflow_node.deploy.form.unicloud_webhost_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormUniCloudWebHostConfig;
