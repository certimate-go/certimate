import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import CodeInput from "@/components/CodeInput";

type NotifyNodeConfigFormWebhookConfigFieldValues = Nullish<{
  webhookData: string;
}>;

export type NotifyNodeConfigFormWebhookConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: NotifyNodeConfigFormWebhookConfigFieldValues;
  onValuesChange?: (values: NotifyNodeConfigFormWebhookConfigFieldValues) => void;
};

const initFormModel = (): NotifyNodeConfigFormWebhookConfigFieldValues => {
  return {};
};

const NotifyNodeConfigFormWebhookConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: NotifyNodeConfigFormWebhookConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    webhookData: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;

        try {
          const obj = JSON.parse(v);
          return typeof obj === "object" && !Array.isArray(obj);
        } catch {
          return false;
        }
      }, t("workflow_node.notify.form.webhook_data.errmsg.json_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const handleWebhookDataBlur = () => {
    const value = formInst.getFieldValue("webhookData");
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue("webhookData", json);
    } catch {
      return;
    }
  };

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
        name="webhookData"
        label={t("workflow_node.notify.form.webhook_data.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.webhook_data.tooltip") }}></span>}
      >
        <CodeInput
          height="auto"
          minHeight="64px"
          maxHeight="256px"
          language="json"
          placeholder={t("workflow_node.notify.form.webhook_data.placeholder")}
          onBlur={handleWebhookDataBlur}
        />
      </Form.Item>

      <Form.Item>
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.webhook_data.guide") }}></span>} />
      </Form.Item>
    </Form>
  );
};

export default NotifyNodeConfigFormWebhookConfig;
