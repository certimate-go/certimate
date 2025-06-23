import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormPushoverFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    token: z
      .string(t("settings.notification.channel.form.pushover_token.placeholder"))
      .nonempty(t("settings.notification.channel.form.pushover_token.placeholder")),
    user: z
      .string(t("settings.notification.channel.form.pushover_user.placeholder"))
      .nonempty(t("settings.notification.channel.form.pushover_user.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="token"
        label={t("settings.notification.channel.form.pushover_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.pushover_token.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.pushover_token.placeholder")} />
      </Form.Item>
      <Form.Item
        name="user"
        label={t("settings.notification.channel.form.pushover_user.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.pushover_user.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.pushover_user.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormPushoverFields;
