import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormMattermostFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    channelId: z
      .string(t("settings.notification.channel.form.mattermost_channel_id.placeholder"))
      .nonempty(t("settings.notification.channel.form.mattermost_channel_id.placeholder")),
    username: z
      .string(t("settings.notification.channel.form.mattermost_username.placeholder"))
      .nonempty(t("settings.notification.channel.form.mattermost_username.placeholder")),
    password: z
      .string(t("settings.notification.channel.form.mattermost_password.placeholder"))
      .nonempty(t("settings.notification.channel.form.mattermost_password.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="serverUrl"
        label={t("settings.notification.channel.form.mattermost_server_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.mattermost_server_url.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.mattermost_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="channelId"
        label={t("settings.notification.channel.form.mattermost_channel_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.mattermost_channel_id.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.mattermost_channel_id.placeholder")} />
      </Form.Item>

      <Form.Item name="username" label={t("settings.notification.channel.form.mattermost_username.label")} rules={[formRule]}>
        <Input placeholder={t("settings.notification.channel.form.mattermost_username.placeholder")} />
      </Form.Item>

      <Form.Item name="password" label={t("settings.notification.channel.form.mattermost_password.label")} rules={[formRule]}>
        <Input.Password placeholder={t("settings.notification.channel.form.mattermost_password.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormMattermostFields;
