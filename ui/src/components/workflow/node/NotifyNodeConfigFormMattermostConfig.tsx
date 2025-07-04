import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type NotifyNodeConfigFormMattermostConfigFieldValues = Nullish<{
  channelId?: string;
}>;

export type NotifyNodeConfigFormMattermostConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: NotifyNodeConfigFormMattermostConfigFieldValues;
  onValuesChange?: (values: NotifyNodeConfigFormMattermostConfigFieldValues) => void;
};

const initFormModel = (): NotifyNodeConfigFormMattermostConfigFieldValues => {
  return {};
};

const NotifyNodeConfigFormMattermostConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: NotifyNodeConfigFormMattermostConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    channelId: z.string().nullish(),
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
        name="channelId"
        label={t("workflow_node.notify.form.mattermost_channel_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.mattermost_channel_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.notify.form.mattermost_channel_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default NotifyNodeConfigFormMattermostConfig;
