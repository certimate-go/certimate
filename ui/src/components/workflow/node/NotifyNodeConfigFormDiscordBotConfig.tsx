import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type NotifyNodeConfigFormDiscordBotConfigFieldValues = Nullish<{
  channelId?: string;
}>;

export type NotifyNodeConfigFormDiscordBotConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: NotifyNodeConfigFormDiscordBotConfigFieldValues;
  onValuesChange?: (values: NotifyNodeConfigFormDiscordBotConfigFieldValues) => void;
};

const initFormModel = (): NotifyNodeConfigFormDiscordBotConfigFieldValues => {
  return {};
};

const NotifyNodeConfigFormDiscordBotConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: NotifyNodeConfigFormDiscordBotConfigProps) => {
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
        label={t("workflow_node.notify.form.discordbot_channel_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.discordbot_channel_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.notify.form.discordbot_channel_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default NotifyNodeConfigFormDiscordBotConfig;
