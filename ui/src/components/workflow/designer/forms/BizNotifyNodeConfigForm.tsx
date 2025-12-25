import { useEffect, useMemo } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconChevronDown, IconPlus } from "@tabler/icons-react";
import { type AnchorProps, Button, Divider, Form, type FormInstance, Input, Switch, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AccessEditDrawer from "@/components/access/AccessEditDrawer";
import AccessSelect from "@/components/access/AccessSelect";
import PresetNotifyTemplatesPopselect from "@/components/preset/PresetNotifyTemplatesPopselect";
import NotificationProviderPicker from "@/components/provider/NotificationProviderPicker";
import NotificationProviderSelect from "@/components/provider/NotificationProviderSelect";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { type AccessModel } from "@/domain/access";
import { notificationProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForBizNotify, defaultNodeConfigForBizNotify } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";

import { FormNestedFieldsContextProvider, NodeFormContextProvider } from "./_context";
import BizNotifyNodeConfigFieldsProvider from "./BizNotifyNodeConfigFieldsProvider";
import { NodeType } from "../nodes/typings";

export interface BizNotifyNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BizNotifyNodeConfigForm = ({ node, ...props }: BizNotifyNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizNotify) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizNotify}`);
  }

  const { i18n, t } = useTranslation();

  const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));
  const accessOptionFilter = (_: string, option: AccessModel) => {
    if (option.reserve !== "notif") return false;
    return notificationProvidersMap.get(fieldProvider)?.provider === option.provider;
  };

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBizNotify | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizNotifyNodeConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldProvider = Form.useWatch("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch("providerAccessId", { form: formInst, preserve: true });

  const renderNestedFieldProviderComponent = BizNotifyNodeConfigFieldsProvider.useComponent(fieldProvider, {});

  useEffect(() => {
    // 如果未选择通知渠道，则清空授权信息
    if (!fieldProvider && fieldProviderAccessId) {
      formInst.setFieldValue("providerAccessId", void 0);
      return;
    }

    // 如果已选择通知渠道只有一个授权信息，则自动选择该授权信息
    if (fieldProvider && !fieldProviderAccessId) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilter(access.provider, access))
        .filter((access) => notificationProvidersMap.get(fieldProvider)?.provider === access.provider);
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("providerAccessId", availableAccesses[0].id);
      }
    }
  }, [fieldProvider, fieldProviderAccessId]);

  const handleProviderPick = (value: string) => {
    formInst.setFieldValue("provider", value);
    formInst.setFieldValue("providerAccessId", void 0);
    formInst.setFieldValue("providerConfig", void 0);
  };

  const handleProviderSelect = (value?: string | undefined) => {
    // 切换通知渠道时重置表单，避免其他通知渠道的配置字段影响当前通知渠道
    if (initialValues?.provider === value) {
      formInst.setFieldValue("providerAccessId", void 0);
      formInst.resetFields(["providerConfig"]);
    } else {
      formInst.setFieldValue("providerAccessId", void 0);
      formInst.setFieldValue("providerConfig", void 0);
    }
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <Show when={!fieldProvider}>
          <NotificationProviderPicker
            autoFocus
            placeholder={t("workflow_node.notify.form.provider.search.placeholder")}
            showAvailability
            showSearch
            onSelect={handleProviderPick}
          />
        </Show>

        <div style={{ display: fieldProvider ? "block" : "none" }}>
          <div id="parameters" data-anchor="parameters">
            <Form.Item name="subject" label={t("workflow_node.notify.form.subject.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.notify.form.subject.placeholder")} />
            </Form.Item>

            <Form.Item label={t("workflow_node.notify.form.message.label")}>
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <PresetNotifyTemplatesPopselect
                  trigger={["click"]}
                  onSelect={(_, template) => {
                    if (template) {
                      formInst.setFieldValue("subject", template.subject);
                      formInst.setFieldValue("message", template.message);
                    }
                  }}
                >
                  <Button size="small" type="link">
                    {t("preset.dropdown.notification.button")}
                    <IconChevronDown size="1.25em" />
                  </Button>
                </PresetNotifyTemplatesPopselect>
              </div>
              <Form.Item name="message" noStyle rules={[formRule]}>
                <Input.TextArea autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("workflow_node.notify.form.message.placeholder")} />
              </Form.Item>
            </Form.Item>

            <Form.Item>
              <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.template.guide") }}></span>} />
            </Form.Item>
          </div>

          <div id="channel" data-anchor="channel">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.notify.form_anchor.channel.title")}
              </Typography.Text>
            </Divider>

            <Form.Item name="provider" label={t("workflow_node.notify.form.provider.label")} rules={[formRule]}>
              <NotificationProviderSelect
                allowClear
                disabled={!!initialValues?.provider}
                placeholder={t("workflow_node.notify.form.provider.placeholder")}
                showAvailability
                showSearch
                onSelect={handleProviderSelect}
                onClear={handleProviderSelect}
              />
            </Form.Item>

            <Form.Item label={t("workflow_node.notify.form.provider_access.label")}>
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <AccessEditDrawer
                  data={{ provider: notificationProvidersMap.get(fieldProvider!)?.provider }}
                  mode="create"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.notify.form.provider_access.button")}
                      <IconPlus size="1.25em" />
                    </Button>
                  }
                  usage="notification"
                  afterSubmit={(record) => {
                    if (!accessOptionFilter(record.provider, record)) return;
                    if (notificationProvidersMap.get(fieldProvider!)?.provider !== record.provider) return;
                    formInst.setFieldValue("providerAccessId", record.id);
                  }}
                />
              </div>
              <Form.Item name="providerAccessId" dependencies={["provider"]} noStyle rules={[formRule]}>
                <AccessSelect
                  disabled={!fieldProvider}
                  placeholder={t("workflow_node.notify.form.provider_access.placeholder")}
                  showSearch
                  onFilter={accessOptionFilter}
                />
              </Form.Item>
            </Form.Item>

            <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
              {renderNestedFieldProviderComponent && <>{renderNestedFieldProviderComponent}</>}
            </FormNestedFieldsContextProvider>
          </div>

          <div id="strategy" data-anchor="strategy">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.notify.form_anchor.strategy.title")}
              </Typography.Text>
            </Divider>

            <Form.Item label={t("workflow_node.notify.form.skip_on_all_prev_skipped.label")}>
              <span className="me-2 inline-block">{t("workflow_node.notify.form.skip_on_all_prev_skipped.prefix")}</span>
              <span className="inline-block">
                <Form.Item name="skipOnAllPrevSkipped" noStyle rules={[formRule]}>
                  <Switch
                    checkedChildren={t("workflow_node.notify.form.skip_on_all_prev_skipped.switch.on")}
                    unCheckedChildren={t("workflow_node.notify.form.skip_on_all_prev_skipped.switch.off")}
                  />
                </Form.Item>
              </span>
              <span className="ms-2 inline-block">{t("workflow_node.notify.form.skip_on_all_prev_skipped.suffix")}</span>
            </Form.Item>
          </div>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters", "channel", "strategy"].map((key) => ({
    key: key,
    title: t(`workflow_node.notify.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    subject: "",
    message: "",
    ...(defaultNodeConfigForBizNotify() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    subject: z.string().nonempty(t("workflow_node.notify.form.subject.placeholder")),
    message: z.string().nonempty(t("workflow_node.notify.form.message.placeholder")),
    provider: z.string().nonempty(t("workflow_node.notify.form.provider.placeholder")),
    providerAccessId: z.string().nonempty(t("workflow_node.notify.form.provider_access.placeholder")),
    providerConfig: z.any().nullish(),
    skipOnAllPrevSkipped: z.boolean().nullish(),
  });
};

const _default = Object.assign(BizNotifyNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
