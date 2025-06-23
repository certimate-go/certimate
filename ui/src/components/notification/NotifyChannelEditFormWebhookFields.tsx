import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormWebhookFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    url: z.url(t("common.errmsg.url_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <div>
      <Form.Item name="url" label={t("settings.notification.channel.form.webhook_url.label")} rules={[formRule]}>
        <Input placeholder={t("settings.notification.channel.form.webhook_url.placeholder")} />
      </Form.Item>
    </div>
  );
};

export default NotifyChannelEditFormWebhookFields;
