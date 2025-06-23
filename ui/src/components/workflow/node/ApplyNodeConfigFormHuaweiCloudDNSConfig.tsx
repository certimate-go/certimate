import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type ApplyNodeConfigFormHuaweiCloudDNSConfigFieldValues = Nullish<{
  region: string;
}>;

export type ApplyNodeConfigFormHuaweiCloudDNSConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: ApplyNodeConfigFormHuaweiCloudDNSConfigFieldValues;
  onValuesChange?: (values: ApplyNodeConfigFormHuaweiCloudDNSConfigFieldValues) => void;
};

const initFormModel = (): ApplyNodeConfigFormHuaweiCloudDNSConfigFieldValues => {
  return {
    region: "cn-north-1",
  };
};

const ApplyNodeConfigFormHuaweiCloudDNSConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: ApplyNodeConfigFormHuaweiCloudDNSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.apply.form.huaweicloud_dns_region.placeholder"))
      .nonempty(t("workflow_node.apply.form.huaweicloud_dns_region.placeholder")),
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
        label={t("workflow_node.apply.form.huaweicloud_dns_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.huaweicloud_dns_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.huaweicloud_dns_region.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default ApplyNodeConfigFormHuaweiCloudDNSConfig;
