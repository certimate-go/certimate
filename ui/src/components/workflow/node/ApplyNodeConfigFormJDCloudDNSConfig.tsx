import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type ApplyNodeConfigFormJDCloudDNSConfigFieldValues = Nullish<{
  regionId: string;
}>;

export type ApplyNodeConfigFormJDCloudDNSConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: ApplyNodeConfigFormJDCloudDNSConfigFieldValues;
  onValuesChange?: (values: ApplyNodeConfigFormJDCloudDNSConfigFieldValues) => void;
};

const initFormModel = (): ApplyNodeConfigFormJDCloudDNSConfigFieldValues => {
  return {
    regionId: "cn-north-1",
  };
};

const ApplyNodeConfigFormJDCloudDNSConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: ApplyNodeConfigFormJDCloudDNSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    regionId: z
      .string(t("workflow_node.apply.form.jdcloud_dns_region_id.placeholder"))
      .nonempty(t("workflow_node.apply.form.jdcloud_dns_region_id.placeholder")),
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
        name="regionId"
        label={t("workflow_node.apply.form.jdcloud_dns_region_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.jdcloud_dns_region_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.jdcloud_dns_region_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default ApplyNodeConfigFormJDCloudDNSConfig;
