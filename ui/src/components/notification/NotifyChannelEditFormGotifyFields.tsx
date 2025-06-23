import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormGotifyFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    url: z.url(t("common.errmsg.url_invalid")),
    token: z.string(t("settings.notification.channel.form.gotify_token.placeholder")),
    priority: z.preprocess((val) => Number(val), z.number().gte(0, t("settings.notification.channel.form.gotify_priority.error.gte0"))),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="url"
        label={t("settings.notification.channel.form.gotify_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.gotify_url.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.gotify_url.placeholder")} />
      </Form.Item>
      <Form.Item
        name="token"
        label={t("settings.notification.channel.form.gotify_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.gotify_token.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.gotify_token.placeholder")} />
      </Form.Item>
      <Form.Item
        name="priority"
        label={t("settings.notification.channel.form.gotify_priority.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.gotify_priority.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.gotify_priority.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormGotifyFields;
