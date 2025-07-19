import { forwardRef, memo, useEffect, useImperativeHandle, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { PlusOutlined as PlusOutlinedIcon } from "@ant-design/icons";
import { Button, Divider, Flex, Form, type FormInstance, Input, Switch, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import AccessEditModal from "@/components/access/AccessEditModal";
import AccessSelect from "@/components/access/AccessSelect";
import NotificationProviderSelect from "@/components/provider/NotificationProviderSelect";
import Show from "@/components/Show";
import { ACCESS_USAGES, NOTIFICATION_PROVIDERS, accessProvidersMap, notificationProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForNotify, defaultNodeConfigForNotify } from "@/domain/workflow";
import { useAntdForm, useAntdFormName, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";

import NotifyNodeConfigFormDiscordBotConfig from "./NotifyNodeConfigFormDiscordBotConfig";
import NotifyNodeConfigFormEmailConfig from "./NotifyNodeConfigFormEmailConfig";
import NotifyNodeConfigFormMattermostConfig from "./NotifyNodeConfigFormMattermostConfig";
import NotifyNodeConfigFormSlackBotConfig from "./NotifyNodeConfigFormSlackBotConfig";
import NotifyNodeConfigFormTelegramBotConfig from "./NotifyNodeConfigFormTelegramBotConfig";
import NotifyNodeConfigFormWebhookConfig from "./NotifyNodeConfigFormWebhookConfig";

type NotifyNodeConfigFormFieldValues = Partial<WorkflowNodeConfigForNotify>;

export type NotifyNodeConfigFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: NotifyNodeConfigFormFieldValues;
  onValuesChange?: (values: NotifyNodeConfigFormFieldValues) => void;
};

export type NotifyNodeConfigFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<NotifyNodeConfigFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<NotifyNodeConfigFormFieldValues>["resetFields"];
  validateFields: FormInstance<NotifyNodeConfigFormFieldValues>["validateFields"];
};

const initFormModel = (): NotifyNodeConfigFormFieldValues => {
  return defaultNodeConfigForNotify();
};

const NotifyNodeConfigForm = forwardRef<NotifyNodeConfigFormInstance, NotifyNodeConfigFormProps>(
  ({ className, style, disabled, initialValues, onValuesChange }, ref) => {
    const { t } = useTranslation();

    const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));

    const formSchema = z.object({
      subject: z
        .string(t("workflow_node.notify.form.subject.placeholder"))
        .min(1, t("workflow_node.notify.form.subject.placeholder"))
        .max(1000, t("common.errmsg.string_max", { max: 1000 })),
      message: z
        .string(t("workflow_node.notify.form.message.placeholder"))
        .min(1, t("workflow_node.notify.form.message.placeholder"))
        .max(1000, t("common.errmsg.string_max", { max: 1000 })),
      channel: z.string().nullish(),
      provider: z.string(t("workflow_node.notify.form.provider.placeholder")).nonempty(t("workflow_node.notify.form.provider.placeholder")),
      providerAccessId: z
        .string(t("workflow_node.notify.form.provider_access.placeholder"))
        .nonempty(t("workflow_node.notify.form.provider_access.placeholder")),
      providerConfig: z.any().nullish(),
      skipOnAllPrevSkipped: z.boolean().nullish(),
    });
    const formRule = createSchemaFieldRule(formSchema);
    const { form: formInst, formProps } = useAntdForm({
      name: "workflowNodeNotifyConfigForm",
      initialValues: initialValues ?? initFormModel(),
    });

    const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });
    const fieldProviderAccessId = Form.useWatch<string>("providerAccessId", formInst);

    const [showProvider, setShowProvider] = useState(false);
    useEffect(() => {
      // 通常情况下每个授权信息只对应一个消息通知提供商，此时无需显示消息通知提供商字段；
      // 如果对应多个，则显示。
      if (fieldProviderAccessId) {
        const access = accesses.find((e) => e.id === fieldProviderAccessId);
        const providers = Array.from(notificationProvidersMap.values()).filter((e) => e.provider === access?.provider);
        setShowProvider(providers.length > 1);
      } else {
        setShowProvider(false);
      }
    }, [accesses, fieldProviderAccessId]);

    const [nestedFormInst] = Form.useForm();
    const nestedFormName = useAntdFormName({ form: nestedFormInst, name: "workflowNodeNotifyConfigFormProviderConfigForm" });
    const nestedFormEl = useMemo(() => {
      const nestedFormProps = {
        form: nestedFormInst,
        formName: nestedFormName,
        disabled: disabled,
        initialValues: initialValues?.providerConfig,
      };

      /*
        注意：如果追加新的子组件，请保持以 ASCII 排序。
        NOTICE: If you add new child component, please keep ASCII order.
       */
      switch (fieldProvider) {
        case NOTIFICATION_PROVIDERS.DISCORDBOT:
          return <NotifyNodeConfigFormDiscordBotConfig {...nestedFormProps} />;
        case NOTIFICATION_PROVIDERS.EMAIL:
          return <NotifyNodeConfigFormEmailConfig {...nestedFormProps} />;
        case NOTIFICATION_PROVIDERS.MATTERMOST:
          return <NotifyNodeConfigFormMattermostConfig {...nestedFormProps} />;
        case NOTIFICATION_PROVIDERS.SLACKBOT:
          return <NotifyNodeConfigFormSlackBotConfig {...nestedFormProps} />;
        case NOTIFICATION_PROVIDERS.TELEGRAMBOT:
          return <NotifyNodeConfigFormTelegramBotConfig {...nestedFormProps} />;
        case NOTIFICATION_PROVIDERS.WEBHOOK:
          return <NotifyNodeConfigFormWebhookConfig {...nestedFormProps} />;
      }
    }, [disabled, initialValues?.providerConfig, fieldProvider, nestedFormInst, nestedFormName]);

    const handleProviderSelect = (value: string) => {
      // 切换消息通知提供商时联动授权信息
      if (initialValues?.provider === value) {
        formInst.setFieldValue("providerAccessId", initialValues?.providerAccessId);
        onValuesChange?.(formInst.getFieldsValue(true));
      } else {
        if (notificationProvidersMap.get(fieldProvider)?.provider !== notificationProvidersMap.get(value)?.provider) {
          formInst.setFieldValue("providerAccessId", undefined);
          onValuesChange?.(formInst.getFieldsValue(true));
        }
      }
    };

    const handleProviderAccessSelect = (value: string) => {
      // 切换授权信息时联动消息通知提供商
      const access = accesses.find((access) => access.id === value);
      const provider = Array.from(notificationProvidersMap.values()).find((provider) => provider.provider === access?.provider);
      if (fieldProvider !== provider?.type) {
        formInst.setFieldValue("provider", provider?.type);
        onValuesChange?.(formInst.getFieldsValue(true));
      }
    };

    const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
      onValuesChange?.(values as NotifyNodeConfigFormFieldValues);
    };

    useImperativeHandle(ref, () => {
      return {
        getFieldsValue: () => {
          const values = formInst.getFieldsValue(true);
          values.providerConfig = nestedFormInst.getFieldsValue();
          return values;
        },
        resetFields: (fields) => {
          formInst.resetFields(fields);

          if (!!fields && fields.includes("providerConfig")) {
            nestedFormInst.resetFields(fields);
          }
        },
        validateFields: (nameList, config) => {
          const t1 = formInst.validateFields(nameList, config);
          const t2 = nestedFormInst.validateFields(undefined, config);
          return Promise.all([t1, t2]).then(() => t1);
        },
      } as NotifyNodeConfigFormInstance;
    });

    return (
      <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
        <Form.Item name="subject" label={t("workflow_node.notify.form.subject.label")} rules={[formRule]}>
          <Input placeholder={t("workflow_node.notify.form.subject.placeholder")} />
        </Form.Item>

        <Form.Item name="message" label={t("workflow_node.notify.form.message.label")} rules={[formRule]}>
          <Input.TextArea autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("workflow_node.notify.form.message.placeholder")} />
        </Form.Item>

        <Form.Item name="provider" label={t("workflow_node.notify.form.provider.label")} hidden={!showProvider} rules={[formRule]}>
          <NotificationProviderSelect
            disabled={!showProvider}
            filter={(record) => {
              if (fieldProviderAccessId) {
                return accesses.find((e) => e.id === fieldProviderAccessId)?.provider === record.provider;
              }

              return true;
            }}
            placeholder={t("workflow_node.notify.form.provider.placeholder")}
            showSearch
            onSelect={handleProviderSelect}
          />
        </Form.Item>

        <Form.Item noStyle>
          <label className="mb-1 block">
            <div className="flex w-full items-center justify-between gap-4">
              <div className="max-w-full grow truncate">
                <span>{t("workflow_node.notify.form.provider_access.label")}</span>
              </div>
              <div className="text-right">
                <AccessEditModal
                  scene="add"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.notify.form.provider_access.button")}
                      <PlusOutlinedIcon className="text-xs" />
                    </Button>
                  }
                  usage="notification-only"
                  afterSubmit={(record) => {
                    const provider = accessProvidersMap.get(record.provider);
                    if (provider?.usages?.includes(ACCESS_USAGES.NOTIFICATION)) {
                      formInst.setFieldValue("providerAccessId", record.id);
                      handleProviderAccessSelect(record.id);
                    }
                  }}
                />
              </div>
            </div>
          </label>
          <Form.Item name="providerAccessId" rules={[formRule]}>
            <AccessSelect
              filter={(record) => {
                if (record.reserve !== "notification") return false;

                const provider = accessProvidersMap.get(record.provider);
                return !!provider?.usages?.includes(ACCESS_USAGES.NOTIFICATION);
              }}
              placeholder={t("workflow_node.notify.form.provider_access.placeholder")}
              showSearch
              onChange={handleProviderAccessSelect}
            />
          </Form.Item>
        </Form.Item>

        <Show when={!!nestedFormEl}>
          <Divider size="small">
            <Typography.Text className="text-xs font-normal" type="secondary">
              {t("workflow_node.notify.form.params_config.label")}
            </Typography.Text>
          </Divider>

          {nestedFormEl}
        </Show>

        <Divider size="small">
          <Typography.Text className="text-xs font-normal" type="secondary">
            {t("workflow_node.notify.form.strategy_config.label")}
          </Typography.Text>
        </Divider>

        <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
          <Form.Item label={t("workflow_node.notify.form.skip_on_all_prev_skipped.label")}>
            <Flex align="center" gap={8} wrap="wrap">
              <div>{t("workflow_node.notify.form.skip_on_all_prev_skipped.prefix")}</div>
              <Form.Item name="skipOnAllPrevSkipped" noStyle rules={[formRule]}>
                <Switch
                  checkedChildren={t("workflow_node.notify.form.skip_on_all_prev_skipped.switch.on")}
                  unCheckedChildren={t("workflow_node.notify.form.skip_on_all_prev_skipped.switch.off")}
                />
              </Form.Item>
              <div>{t("workflow_node.notify.form.skip_on_all_prev_skipped.suffix")}</div>
            </Flex>
          </Form.Item>
        </Form>
      </Form>
    );
  }
);

export default memo(NotifyNodeConfigForm);
