import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormPushPlusFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    token: z.string(t("settings.notification.channel.form.pushplus_token.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="token"
        label={t("settings.notification.channel.form.pushplus_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.pushplus_token.tooltip") }}></span>}
      >
        <Input placeholder={t("settings.notification.channel.form.pushplus_token.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormPushPlusFields;
