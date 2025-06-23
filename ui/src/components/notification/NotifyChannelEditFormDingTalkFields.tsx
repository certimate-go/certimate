import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

const NotifyChannelEditFormDingTalkFields = () => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessToken: z
      .string(t("settings.notification.channel.form.dingtalk_access_token.placeholder"))
      .nonempty(t("settings.notification.channel.form.dingtalk_access_token.placeholder")),
    secret: z
      .string(t("settings.notification.channel.form.dingtalk_secret.placeholder"))
      .nonempty(t("settings.notification.channel.form.dingtalk_secret.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  return (
    <>
      <Form.Item
        name="accessToken"
        label={t("settings.notification.channel.form.dingtalk_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.dingtalk_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("settings.notification.channel.form.dingtalk_access_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secret"
        label={t("settings.notification.channel.form.dingtalk_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("settings.notification.channel.form.dingtalk_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("settings.notification.channel.form.dingtalk_secret.placeholder")} />
      </Form.Item>
    </>
  );
};

export default NotifyChannelEditFormDingTalkFields;
