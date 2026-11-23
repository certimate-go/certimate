import { useEffect, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconPlus } from "@tabler/icons-react";
import { type AnchorProps, Button, Divider, Form, type FormInstance, InputNumber, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import AccessEditDrawer from "@/components/access/AccessEditDrawer";
import AccessSelect from "@/components/access/AccessSelect";
import PurgeProviderPicker from "@/components/provider/PurgeProviderPicker";
import PurgeProviderSelect from "@/components/provider/PurgeProviderSelect";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { type AccessModel } from "@/domain/access";
import { getPurgeProvider, normalizePurgeProviderType, resolvePurgeProvider } from "@/domain/purgeProvider";
import { type WorkflowNodeConfigForBizPurge, defaultNodeConfigForBizPurge } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";

import { FormNestedFieldsContextProvider, NodeFormContextProvider } from "./_context";
import BizPurgeNodeConfigFieldsProvider from "./BizPurgeNodeConfigFieldsProvider";
import { NodeType } from "../nodes/typings";

export interface BizPurgeNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BizPurgeNodeConfigForm = ({ node, ...props }: BizPurgeNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizPurge) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizPurge}`);
  }

  const { i18n, t } = useTranslation();

  const pluginCatalogVersion = usePluginCatalogStore((s) => s.version);

  const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));
  const accessOptionFilter = (_: string, option: AccessModel) => {
    if (option.reserve) return false;
    return option.provider === getPurgeProvider(fieldProvider)?.provider;
  };

  const initialValues = node.form?.getValueIn("config") as WorkflowNodeConfigForBizPurge | undefined;

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizPurgeConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldProvider = Form.useWatch("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch("providerAccessId", { form: formInst, preserve: true });

  const [providerSchemaUnavailable, setProviderSchemaUnavailable] = useState(false);

  useEffect(() => {
    let cancelled = false;
    if (!fieldProvider || getPurgeProvider(fieldProvider)?.source !== "plugin") {
      setProviderSchemaUnavailable(false);
      return;
    }

    setProviderSchemaUnavailable(true);
    void getProviderSchema(resolvePurgeProvider(fieldProvider), "purge").then((envelope: ProviderSchemaEnvelope | null) => {
      if (!cancelled) {
        setProviderSchemaUnavailable(envelope == null);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [fieldProvider, pluginCatalogVersion]);

  useEffect(() => {
    if (!fieldProvider) {
      return;
    }

    const normalized = normalizePurgeProviderType(fieldProvider);
    if (normalized !== fieldProvider) {
      formInst.setFieldValue("provider", normalized);
    }
  }, [fieldProvider, pluginCatalogVersion]);

  const renderNestedFieldProviderComponent = BizPurgeNodeConfigFieldsProvider.useComponent(fieldProvider, {});

  const showProviderAccess = (() => {
    // 内置的清除提供商无需显示授权信息字段
    if (fieldProvider) {
      const provider = getPurgeProvider(fieldProvider);
      return !provider?.builtin;
    }

    return false;
  })();

  useEffect(() => {
    // 如果未选择提供商，则清空授权信息
    if (!fieldProvider && fieldProviderAccessId) {
      formInst.setFieldValue("providerAccessId", void 0);
      return;
    }

    // 如果已选择提供商只有一个授权信息，则自动选择该授权信息
    if (fieldProvider && !fieldProviderAccessId) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilter(access.provider, access))
        .filter((access) => getPurgeProvider(fieldProvider)?.provider === access.provider);
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("providerAccessId", availableAccesses[0].id);
      }
    }
  }, [fieldProvider, fieldProviderAccessId]);

  const handleProviderPick = (value: string) => {
    // 首次选择提供商时重置表单，避免来自导入的配置字段残留
    formInst.setFieldValue("provider", value);
    formInst.setFieldValue("providerAccessId", void 0);
    formInst.setFieldValue("providerConfig", void 0);
  };

  const handleProviderSelect = (value?: string | undefined) => {
    // 切换提供商时重置表单，避免其他提供商的配置字段残留
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
          <PurgeProviderPicker
            placeholder={t("workflow_node.purge.form.provider.search.placeholder")}
            showAvailability
            showSearch
            onSelect={handleProviderPick}
          />
        </Show>

        <div style={{ display: fieldProvider ? "block" : "none" }}>
          <div id="parameters" data-anchor="parameters">
            <Form.Item>
              <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.purge.form.guide") }}></span>} />
            </Form.Item>
          </div>

          <div id="purging" data-anchor="purging">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.purge.form_anchor.purging.title")}
              </Typography.Text>
            </Divider>

            <Form.Item
              name="provider"
              label={t("workflow_node.purge.form.provider.label")}
              rules={[formRule, purgeSchemaUnavailableRule(t)("workflow_node.purge.form.provider.unavailable")(providerSchemaUnavailable)]}
            >
              <PurgeProviderSelect
                allowClear
                disabled={!!initialValues?.provider}
                placeholder={t("workflow_node.purge.form.provider.placeholder")}
                showAvailability
                showSearch
                onSelect={handleProviderSelect}
                onClear={handleProviderSelect}
              />
            </Form.Item>

            <Form.Item className="relative" hidden={!showProviderAccess} label={t("workflow_node.purge.form.provider_access.label")}>
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <AccessEditDrawer
                  data={{ provider: getPurgeProvider(fieldProvider)?.provider }}
                  mode="create"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.purge.form.provider_access.button")}
                      <IconPlus size="1.25em" />
                    </Button>
                  }
                  usage="hosting"
                  afterSubmit={(record) => {
                    if (!accessOptionFilter(record.provider, record)) return;
                    if (getPurgeProvider(fieldProvider)?.provider !== record.provider) return;
                    formInst.setFieldValue("providerAccessId", record.id);
                  }}
                />
              </div>
              <Form.Item name="providerAccessId" dependencies={["provider"]} rules={[formRule]} noStyle>
                <AccessSelect
                  disabled={!fieldProvider}
                  placeholder={t("workflow_node.purge.form.provider_access.placeholder")}
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
                {t("workflow_node.purge.form_anchor.strategy.title")}
              </Typography.Text>
            </Divider>

            <Form.Item label={t("workflow_node.purge.form.expired_days.label")}>
              <span className="me-2 inline-block">{t("workflow_node.purge.form.expired_days.prefix")}</span>
              <span className="inline-block">
                <Form.Item name="expiredDays" noStyle rules={[formRule]}>
                  <InputNumber
                    className="w-24"
                    min={0}
                    max={365}
                    placeholder={t("workflow_node.purge.form.expired_days.placeholder")}
                    suffix={t("workflow_node.purge.form.expired_days.unit")}
                  />
                </Form.Item>
              </span>
              <span className="ms-2 inline-block">{t("workflow_node.purge.form.expired_days.suffix")}</span>
            </Form.Item>
          </div>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters", "purging", "strategy"].map((key) => ({
    key: key,
    title: t(`workflow_node.purge.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    ...(defaultNodeConfigForBizPurge() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

export const purgeSchemaUnavailableRule =
  (t: (key: string) => string) =>
  (messageKey: string) =>
  (unavailable: boolean): { validator: () => Promise<void> } => ({
    validator: () => (unavailable ? Promise.reject(new Error(t(messageKey))) : Promise.resolve()),
  });

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      provider: z.string().nonempty(),
      providerAccessId: z.string().nullish(),
      providerConfig: z.any().nullish(),
      expiredDays: z.coerce.number().int().gte(0).or(z.literal("")).nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.provider) {
        const provider = getPurgeProvider(values.provider);
        if (!provider) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.purge.form.provider.unavailable"),
            path: ["provider"],
          });
        } else if (!provider.builtin && !values.providerAccessId) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.purge.form.provider_access.placeholder"),
            path: ["providerAccessId"],
          });
        }
      }
    });
};

const _default = Object.assign(BizPurgeNodeConfigForm, {
  getAnchorItems,
  getSchema,
  purgeSchemaUnavailableRule,
});

export default _default;
