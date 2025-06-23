import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type ApplyNodeConfigFormAliyunESAConfigFieldValues = Nullish<{
  region: string;
}>;

export type ApplyNodeConfigFormAliyunESAConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: ApplyNodeConfigFormAliyunESAConfigFieldValues;
  onValuesChange?: (values: ApplyNodeConfigFormAliyunESAConfigFieldValues) => void;
};

const initFormModel = (): ApplyNodeConfigFormAliyunESAConfigFieldValues => {
  return {};
};

const ApplyNodeConfigFormAliyunESAConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: ApplyNodeConfigFormAliyunESAConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z.string(t("workflow_node.apply.form.aliyun_esa_region.placeholder")).nonempty(t("workflow_node.apply.form.aliyun_esa_region.placeholder")),
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
        label={t("workflow_node.apply.form.aliyun_esa_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.aliyun_esa_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.aliyun_esa_region.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default ApplyNodeConfigFormAliyunESAConfig;
