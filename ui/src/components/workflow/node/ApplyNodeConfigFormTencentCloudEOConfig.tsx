import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type ApplyNodeConfigFormTencentCloudEOConfigFieldValues = Nullish<{
  zoneId: string;
}>;

export type ApplyNodeConfigFormTencentCloudEOConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: ApplyNodeConfigFormTencentCloudEOConfigFieldValues;
  onValuesChange?: (values: ApplyNodeConfigFormTencentCloudEOConfigFieldValues) => void;
};

const initFormModel = (): ApplyNodeConfigFormTencentCloudEOConfigFieldValues => {
  return {};
};

const ApplyNodeConfigFormTencentCloudEOConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: ApplyNodeConfigFormTencentCloudEOConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    zoneId: z
      .string(t("workflow_node.apply.form.tencentcloud_eo_zone_id.placeholder"))
      .nonempty(t("workflow_node.apply.form.tencentcloud_eo_zone_id.placeholder")),
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
        name="zoneId"
        label={t("workflow_node.apply.form.tencentcloud_eo_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.tencentcloud_eo_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.tencentcloud_eo_zone_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default ApplyNodeConfigFormTencentCloudEOConfig;
