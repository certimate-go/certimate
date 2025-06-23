import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useRequest } from "ahooks";
import { Button, Form, Input, Skeleton, message, notification } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { ClientResponseError } from "pocketbase";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { type NotifyTemplatesSettingsContent, SETTINGS_NAMES, defaultNotifyTemplate } from "@/domain/settings";
import { useAntdForm } from "@/hooks";
import { get as getSettings, save as saveSettings } from "@/repository/settings";
import { getErrMsg } from "@/utils/error";

export type NotifyTemplateFormProps = {
  className?: string;
  style?: React.CSSProperties;
};

const NotifyTemplateForm = ({ className, style }: NotifyTemplateFormProps) => {
  const { t } = useTranslation();

  const [messageApi, MessageContextHolder] = message.useMessage();
  const [notificationApi, NotificationContextHolder] = notification.useNotification();

  const formSchema = z.object({
    subject: z
      .string()
      .min(1, t("settings.notification.template.form.subject.placeholder"))
      .max(1000, t("common.errmsg.string_max", { max: 1000 })),
    message: z
      .string()
      .min(1, t("settings.notification.template.form.message.placeholder"))
      .max(1000, t("common.errmsg.string_max", { max: 1000 })),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: defaultNotifyTemplate,
    onSubmit: async (values) => {
      try {
        const settings = await getSettings<NotifyTemplatesSettingsContent>(SETTINGS_NAMES.NOTIFY_TEMPLATES);
        await saveSettings<NotifyTemplatesSettingsContent>({
          ...settings,
          content: {
            notifyTemplates: [values],
          },
        });

        messageApi.success(t("common.text.operation_succeeded"));
      } catch (err) {
        notificationApi.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const { loading } = useRequest(
    () => {
      return getSettings<NotifyTemplatesSettingsContent>(SETTINGS_NAMES.NOTIFY_TEMPLATES);
    },
    {
      onError: (err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);
      },
      onFinally: (_, resp) => {
        const template = resp?.content?.notifyTemplates?.[0] ?? defaultNotifyTemplate;
        formInst.setFieldsValue(template);
      },
    }
  );

  const handleInputChange = () => {
    setFormChanged(true);
  };

  return (
    <div className={className} style={style}>
      {MessageContextHolder}
      {NotificationContextHolder}

      <Show when={!loading} fallback={<Skeleton active />}>
        <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
          <Form.Item
            name="subject"
            label={t("settings.notification.template.form.subject.label")}
            extra={t("settings.notification.template.form.subject.extra")}
            rules={[formRule]}
          >
            <Input placeholder={t("settings.notification.template.form.subject.placeholder")} onChange={handleInputChange} />
          </Form.Item>

          <Form.Item
            name="message"
            label={t("settings.notification.template.form.message.label")}
            extra={t("settings.notification.template.form.message.extra")}
            rules={[formRule]}
          >
            <Input.TextArea
              autoSize={{ minRows: 3, maxRows: 10 }}
              placeholder={t("settings.notification.template.form.message.placeholder")}
              onChange={handleInputChange}
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
              {t("common.button.save")}
            </Button>
          </Form.Item>
        </Form>
      </Show>
    </div>
  );
};

export default NotifyTemplateForm;
