import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormTelegramFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiToken: z
      .string(t("settings.notification.channel.form.telegram_api_token.placeholder"))
      .min(1, t("settings.notification.channel.form.telegram_api_token.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    chatId: z
      .string(t("settings.notification.channel.form.telegram_chat_id.placeholder"))
      .min(1, t("settings.notification.channel.form.telegram_chat_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="apiToken"
        label={t("settings.notification.channel.form.telegram_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.telegram_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("settings.notification.channel.form.telegram_api_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name="chatId"
        label={t("settings.notification.channel.form.telegram_chat_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.telegram_chat_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("settings.notification.channel.form.telegram_chat_id.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormTelegramFields;
