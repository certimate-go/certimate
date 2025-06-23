import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Form, InputNumber, Skeleton, message, notification } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { produce } from "immer";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { type PersistenceSettingsContent, SETTINGS_NAMES, type SettingsModel } from "@/domain/settings";
import { useAntdForm } from "@/hooks";
import { get as getSettings, save as saveSettings } from "@/repository/settings";
import { getErrMsg } from "@/utils/error";

const SettingsPersistence = () => {
  const { t } = useTranslation();

  const [messageApi, MessageContextHolder] = message.useMessage();
  const [notificationApi, NotificationContextHolder] = notification.useNotification();

  const [settings, setSettings] = useState<SettingsModel<PersistenceSettingsContent>>();
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);

      const settings = await getSettings<PersistenceSettingsContent>(SETTINGS_NAMES.PERSISTENCE);
      setSettings(settings);

      setLoading(false);
    };

    fetchData();
  }, []);

  const formSchema = z.object({
    workflowRunsMaxDaysRetention: z
      .number(t("settings.persistence.form.workflow_runs_max_days.placeholder"))
      .gte(0, t("settings.persistence.form.workflow_runs_max_days.placeholder")),
    expiredCertificatesMaxDaysRetention: z
      .number(t("settings.persistence.form.expired_certificates_max_days.placeholder"))
      .gte(0, t("settings.persistence.form.expired_certificates_max_days.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      workflowRunsMaxDaysRetention: settings?.content?.workflowRunsMaxDaysRetention ?? 0,
      expiredCertificatesMaxDaysRetention: settings?.content?.expiredCertificatesMaxDaysRetention ?? 0,
    },
    onSubmit: async (values) => {
      try {
        await saveSettings(
          produce(settings!, (draft) => {
            draft.content ??= {} as PersistenceSettingsContent;
            draft.content.workflowRunsMaxDaysRetention = values.workflowRunsMaxDaysRetention;
            draft.content.expiredCertificatesMaxDaysRetention = values.expiredCertificatesMaxDaysRetention;
          })
        );

        messageApi.success(t("common.text.operation_succeeded"));
      } catch (err) {
        notificationApi.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const handleInputChange = () => {
    const changed =
      formInst.getFieldValue("workflowRunsMaxDaysRetention") !== formProps.initialValues?.workflowRunsMaxDaysRetention ||
      formInst.getFieldValue("expiredCertificatesMaxDaysRetention") !== formProps.initialValues?.workflowRunsMaxDaysRetention;
    setFormChanged(changed);
  };

  return (
    <>
      {MessageContextHolder}
      {NotificationContextHolder}

      <Show when={!loading} fallback={<Skeleton active />}>
        <div className="md:max-w-[40rem]">
          <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
            <Form.Item
              name="workflowRunsMaxDaysRetention"
              label={t("settings.persistence.form.workflow_runs_max_days.label")}
              extra={<span dangerouslySetInnerHTML={{ __html: t("settings.persistence.form.workflow_runs_max_days.extra") }}></span>}
              rules={[formRule]}
            >
              <InputNumber
                className="w-full"
                min={0}
                max={36500}
                placeholder={t("settings.persistence.form.workflow_runs_max_days.placeholder")}
                addonAfter={t("settings.persistence.form.workflow_runs_max_days.unit")}
                onChange={handleInputChange}
              />
            </Form.Item>

            <Form.Item
              name="expiredCertificatesMaxDaysRetention"
              label={t("settings.persistence.form.expired_certificates_max_days.label")}
              extra={<span dangerouslySetInnerHTML={{ __html: t("settings.persistence.form.expired_certificates_max_days.extra") }}></span>}
              rules={[formRule]}
            >
              <InputNumber
                className="w-full"
                min={0}
                max={36500}
                placeholder={t("settings.persistence.form.expired_certificates_max_days.placeholder")}
                addonAfter={t("settings.persistence.form.expired_certificates_max_days.unit")}
                onChange={handleInputChange}
              />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
                {t("common.button.save")}
              </Button>
            </Form.Item>
          </Form>
        </div>
      </Show>
    </>
  );
};

export default SettingsPersistence;
