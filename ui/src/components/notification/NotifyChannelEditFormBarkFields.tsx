import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormBarkFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")).nullish(),
    deviceKey: z
      .string(t("settings.notification.channel.form.bark_device_key.placeholder"))
      .nonempty(t("settings.notification.channel.form.bark_device_key.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="serverUrl"
        label={t("settings.notification.channel.form.bark_server_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.bark_server_url.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.bark_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="deviceKey"
        label={t("settings.notification.channel.form.bark_device_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.bark_device_key.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.bark_device_key.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormBarkFields;
