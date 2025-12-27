import { memo, useEffect, useMemo, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { Link } from "react-router";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconArrowRight, IconChevronRight, IconCircleMinus, IconMapPin, IconPlus, IconWorldWww } from "@tabler/icons-react";
import { useControllableValue, useMount } from "ahooks";
import {
  type AnchorProps,
  AutoComplete,
  Avatar,
  Button,
  Card,
  Divider,
  Form,
  type FormInstance,
  Input,
  InputNumber,
  Radio,
  Select,
  Space,
  Switch,
  Typography,
} from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AccessEditDrawer from "@/components/access/AccessEditDrawer";
import AccessSelect from "@/components/access/AccessSelect";
import FileTextInput from "@/components/FileTextInput";
import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import ACMEDns01ProviderSelect from "@/components/provider/ACMEDns01ProviderSelect";
import ACMEHttp01ProviderSelect from "@/components/provider/ACMEHttp01ProviderSelect";
import CAProviderSelect from "@/components/provider/CAProviderSelect";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { CA_PROVIDERS, acmeDns01ProvidersMap, acmeHttp01ProvidersMap, caProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForBizApply, defaultNodeConfigForBizApply } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { useContactEmailsStore } from "@/stores/settings";
import { mergeCls } from "@/utils/css";
import { matchSearchOption } from "@/utils/search";
import { isDomain, isHostname, isIPv4, isIPv6 } from "@/utils/validator";
import { getPrivateKeyAlgorithm as getPKIXPrivateKeyAlgorithm, validatePEMPrivateKey } from "@/utils/x509";

import { FormNestedFieldsContextProvider, NodeFormContextProvider } from "./_context";
import BizApplyNodeConfigFieldsProvider from "./BizApplyNodeConfigFieldsProvider";
import { NodeType } from "../nodes/typings";

const MULTIPLE_INPUT_SEPARATOR = ";";

const IDENTIFIER_DOMAIN = "domain" as const;
const IDENTIFIER_IP = "ip" as const;

const CHALLENGE_TYPE_DNS01 = "dns-01" as const;
const CHALLENGE_TYPE_HTTP01 = "http-01" as const;

const KEY_SOURCE_AUTO = "auto" as const;
const KEY_SOURCE_REUSE = "reuse" as const;
const KEY_SOURCE_CUSTOM = "custom" as const;

export interface BizApplyNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BizApplyNodeConfigForm = ({ node, ...props }: BizApplyNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizApply) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizApply}`);
  }

  const { i18n, t } = useTranslation();

  const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));
  const accessOptionFilter = (_: string, option: AccessModel) => {
    if (option.reserve) return false;
    if (fieldChallengeType === CHALLENGE_TYPE_DNS01) return acmeDns01ProvidersMap.get(fieldProvider)?.provider === option.provider;
    if (fieldChallengeType === CHALLENGE_TYPE_HTTP01) return acmeHttp01ProvidersMap.get(fieldProvider)?.provider === option.provider;
    return false;
  };
  const accessOptionFilterForCA = (_: string, option: AccessModel) => {
    if (option.reserve !== "ca") return false;
    return caProvidersMap.get(fieldCAProvider!)?.provider === option.provider;
  };

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBizApply | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  type FormSchema = z.infer<typeof formSchema>;
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<FormSchema>({
    form: props.form,
    name: "workflowNodeBizApplyConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldIdentifier = Form.useWatch("identifier", { form: formInst, preserve: true });
  const fieldChallengeType = Form.useWatch("challengeType", { form: formInst, preserve: true });
  const fieldProvider = Form.useWatch("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch("providerAccessId", { form: formInst, preserve: true });
  const fieldKeySource = Form.useWatch("keySource", { form: formInst, preserve: true });
  const fieldCAProvider = Form.useWatch("caProvider", { form: formInst, preserve: true });
  const fieldCAProviderAccessId = Form.useWatch("caProviderAccessId", { form: formInst, preserve: true });

  const renderNestedFieldProviderComponent = BizApplyNodeConfigFieldsProvider.useComponent(fieldChallengeType, fieldProvider, {});

  const resetFieldIfInvalid = (field: keyof FormSchema) => {
    const fieldSchame = formSchema.pick({ [field]: true });
    const fieldValue = formInst.getFieldValue(field);
    if (!fieldSchame.safeParse({ [field]: fieldValue }).success) {
      formInst.setFieldValue(field, void 0);
    }
  };

  const showProviderAccess = useMemo(() => {
    // 内置的质询提供商（如本地主机）无需显示授权信息字段
    switch (fieldChallengeType) {
      case CHALLENGE_TYPE_DNS01:
        {
          if (fieldProvider) {
            const provider = acmeDns01ProvidersMap.get(fieldProvider);
            return !provider?.builtin;
          }
        }
        break;

      case CHALLENGE_TYPE_HTTP01:
        {
          if (fieldProvider) {
            const provider = acmeHttp01ProvidersMap.get(fieldProvider);
            return !provider?.builtin;
          }
        }
        break;
    }

    return false;
  }, [fieldChallengeType, fieldProvider]);

  const showCAProviderAccess = useMemo(() => {
    // 内置的 CA 提供商（如 Let's Encrypt）无需显示授权信息字段
    if (fieldCAProvider) {
      const provider = caProvidersMap.get(fieldCAProvider);
      return !provider?.builtin;
    }

    return false;
  }, [fieldCAProvider]);

  useEffect(() => {
    // 如果未选择质询提供商，则清空授权信息
    if (!fieldProvider && fieldProviderAccessId) {
      formInst.setFieldValue("providerAccessId", void 0);
      return;
    }

    // 如果已选择质询提供商只有一个授权信息，则自动选择该授权信息
    if (fieldProvider && !fieldProviderAccessId) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilter(access.provider, access))
        .filter((access) => {
          if (fieldChallengeType === CHALLENGE_TYPE_DNS01) return acmeDns01ProvidersMap.get(fieldProvider)?.provider === access.provider;
          if (fieldChallengeType === CHALLENGE_TYPE_HTTP01) return acmeHttp01ProvidersMap.get(fieldProvider)?.provider === access.provider;
          return false;
        });
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("providerAccessId", availableAccesses[0].id);
      }
    }
  }, [fieldChallengeType, fieldProvider, fieldProviderAccessId]);

  useEffect(() => {
    // 如果未选择 CA 提供商，则清空授权信息
    if (!fieldCAProvider && fieldCAProviderAccessId) {
      formInst.setFieldValue("caProviderAccessId", void 0);
      return;
    }

    // 如果已选择 CA 提供商只有一个授权信息，则自动选择该授权信息
    if (fieldCAProvider && !fieldCAProviderAccessId) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilterForCA(access.provider, access))
        .filter((access) => caProvidersMap.get(fieldCAProvider)?.provider === access.provider);
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("caProviderAccessId", availableAccesses[0].id);
      }
    }
  }, [fieldCAProvider, fieldCAProviderAccessId]);

  const handleIdentifierPick = (value: string) => {
    switch (value) {
      case IDENTIFIER_DOMAIN:
        {
          formInst.setFieldValue("identifier", IDENTIFIER_DOMAIN);
          formInst.setFieldValue("domains", formInst.getFieldValue("domains") || "");
          formInst.setFieldValue("challengeType", CHALLENGE_TYPE_DNS01);
        }
        break;

      case IDENTIFIER_IP:
        {
          formInst.setFieldValue("identifier", IDENTIFIER_IP);
          formInst.setFieldValue("ipaddrs", formInst.getFieldValue("ipaddrs") || "");
          formInst.setFieldValue("challengeType", CHALLENGE_TYPE_HTTP01);
          formInst.setFieldValue("caProvider", CA_PROVIDERS.LETSENCRYPT);
          formInst.setFieldValue("caProviderAccessId", void 0);
          formInst.setFieldValue("caProviderConfig", void 0);
          formInst.setFieldValue("acmeProfile", "shortlived");
          formInst.setFieldValue("disableCommonName", true);
          formInst.setFieldValue("skipBeforeExpiryDays", 3);
        }
        break;
    }

    setTimeout(() => handleIdentifierChange(value), 0);
  };

  const handleIdentifierChange = (value: string) => {
    switch (value) {
      case IDENTIFIER_DOMAIN:
        {
          formInst.setFieldValue("ipaddrs", void 0);
        }
        break;

      case IDENTIFIER_IP:
        {
          formInst.setFieldValue("domains", void 0);

          resetFieldIfInvalid("nameservers");
        }
        break;
    }
  };

  const handleChallengeTypeChange = (value: string) => {
    switch (value) {
      case CHALLENGE_TYPE_DNS01:
        {
          formInst.setFieldValue("provider", void 0);
          formInst.setFieldValue("providerAccessId", void 0);
          formInst.setFieldValue("providerConfig", void 0);

          resetFieldIfInvalid("httpDelayWait");
        }
        break;

      case CHALLENGE_TYPE_HTTP01:
        {
          formInst.setFieldValue("provider", void 0);
          formInst.setFieldValue("providerAccessId", void 0);
          formInst.setFieldValue("providerConfig", void 0);

          resetFieldIfInvalid("dnsPropagationWait");
          resetFieldIfInvalid("dnsPropagationTimeout");
          resetFieldIfInvalid("dnsTTL");
        }
        break;
    }
  };

  const handleProviderSelect = (value?: string | undefined) => {
    // 切换质询提供商时重置表单，避免其他提供商的配置字段影响当前提供商
    if (initialValues?.provider === value) {
      formInst.setFieldValue("providerAccessId", void 0);
      formInst.resetFields(["providerConfig"]);
    } else {
      formInst.setFieldValue("providerAccessId", void 0);
      formInst.setFieldValue("providerConfig", void 0);
    }
  };

  const handleKeySourceChange = (value: string) => {
    if (value === initialValues?.keySource) {
      formInst.resetFields(["keyContent"]);
    } else {
      setTimeout(() => {
        formInst.setFieldValue("keyContent", "");
      }, 0);
    }
  };

  const handleCAProviderSelect = (value?: string | undefined) => {
    // 切换 CA 提供商时联动授权信息
    if (value == null || value === "") {
      formInst.setFieldValue("caProvider", void 0);
      formInst.setFieldValue("caProviderAccessId", void 0);
    } else if (value === initialValues?.caProvider) {
      formInst.setFieldValue("caProviderAccessId", initialValues?.caProviderAccessId);
    } else {
      if (caProvidersMap.get(fieldCAProvider!)?.provider !== caProvidersMap.get(value!)?.provider) {
        formInst.setFieldValue("caProviderAccessId", void 0);
      }
    }
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <Show when={!fieldIdentifier}>
          <InternalIdentifierPicker onSelect={handleIdentifierPick} />
        </Show>

        <div style={{ display: fieldIdentifier ? "block" : "none" }}>
          <div id="parameters" data-anchor="parameters">
            <Form.Item name="identifier" hidden label={t("workflow_node.apply.form.identifier.label")} rules={[formRule]}>
              <Radio.Group block onChange={(e) => handleIdentifierChange(e.target.value)}>
                <Radio.Button value={IDENTIFIER_DOMAIN}>{t("workflow_node.apply.form.identifier.option.domain.label")}</Radio.Button>
                <Radio.Button value={IDENTIFIER_IP}>{t("workflow_node.apply.form.identifier.option.ip.label")}</Radio.Button>
              </Radio.Group>
            </Form.Item>

            <Show>
              <Show.Case when={fieldIdentifier === IDENTIFIER_DOMAIN}>
                <Form.Item
                  name="domains"
                  dependencies={["identifier", "challengeType"]}
                  label={t("workflow_node.apply.form.domains.label")}
                  extra={
                    <span
                      dangerouslySetInnerHTML={{
                        __html:
                          fieldChallengeType === CHALLENGE_TYPE_HTTP01
                            ? t("workflow_node.apply.form.domains.help_no_wildcard")
                            : t("workflow_node.apply.form.domains.help"),
                      }}
                    ></span>
                  }
                  rules={[formRule]}
                >
                  <MultipleSplitValueInput
                    modalTitle={t("workflow_node.apply.form.domains.multiple_input_modal.title")}
                    placeholder={t("workflow_node.apply.form.domains.placeholder")}
                    placeholderInModal={t("workflow_node.apply.form.domains.multiple_input_modal.placeholder")}
                    separator={MULTIPLE_INPUT_SEPARATOR}
                    splitOptions={{ removeEmpty: true, trimSpace: true }}
                  />
                </Form.Item>
              </Show.Case>
              <Show.Case when={fieldIdentifier === IDENTIFIER_IP}>
                <Form.Item
                  name="ipaddrs"
                  dependencies={["identifier", "challengeType"]}
                  label={t("workflow_node.apply.form.ipaddrs.label")}
                  extra={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.ipaddrs.help") }}></span>}
                  rules={[formRule]}
                >
                  <MultipleSplitValueInput
                    modalTitle={t("workflow_node.apply.form.ipaddrs.multiple_input_modal.title")}
                    placeholder={t("workflow_node.apply.form.ipaddrs.placeholder")}
                    placeholderInModal={t("workflow_node.apply.form.ipaddrs.multiple_input_modal.placeholder")}
                    separator={MULTIPLE_INPUT_SEPARATOR}
                    splitOptions={{ removeEmpty: true, trimSpace: true }}
                  />
                </Form.Item>
              </Show.Case>
            </Show>

            <Form.Item
              name="contactEmail"
              label={t("workflow_node.apply.form.contact_email.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.contact_email.tooltip") }}></span>}
            >
              <InternalEmailInput />
            </Form.Item>
          </div>

          <div id="challenge" data-anchor="challenge">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.challenge.title")}
              </Typography.Text>
            </Divider>

            <Form.Item
              name="challengeType"
              dependencies={["identifier", "domains", "ipaddrs"]}
              label={t("workflow_node.apply.form.challenge_type.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.challenge_type.tooltip") }}></span>}
            >
              <Radio.Group block onChange={(e) => handleChallengeTypeChange(e.target.value)}>
                <Radio.Button disabled={fieldIdentifier === IDENTIFIER_IP} value={CHALLENGE_TYPE_DNS01}>
                  DNS-01
                </Radio.Button>
                <Radio.Button value={CHALLENGE_TYPE_HTTP01}>HTTP-01</Radio.Button>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              name="provider"
              dependencies={["challengeType"]}
              label={
                fieldChallengeType === CHALLENGE_TYPE_DNS01
                  ? t("workflow_node.apply.form.provider_dns01.label")
                  : fieldChallengeType === CHALLENGE_TYPE_HTTP01
                    ? t("workflow_node.apply.form.provider_http01.label")
                    : t("workflow_node.apply.form.provider.label")
              }
              rules={[formRule]}
            >
              {fieldChallengeType === CHALLENGE_TYPE_DNS01 ? (
                <ACMEDns01ProviderSelect
                  placeholder={t("workflow_node.apply.form.provider_dns01.placeholder")}
                  showAvailability
                  showSearch
                  onSelect={handleProviderSelect}
                  onClear={handleProviderSelect}
                />
              ) : fieldChallengeType === CHALLENGE_TYPE_HTTP01 ? (
                <ACMEHttp01ProviderSelect
                  placeholder={t("workflow_node.apply.form.provider_http01.placeholder")}
                  showAvailability
                  showSearch
                  onSelect={handleProviderSelect}
                  onClear={handleProviderSelect}
                />
              ) : (
                <Select disabled placeholder={t("workflow_node.apply.form.provider.placeholder")} />
              )}
            </Form.Item>

            <Form.Item
              className="relative"
              hidden={!showProviderAccess}
              label={
                fieldChallengeType === CHALLENGE_TYPE_DNS01
                  ? t("workflow_node.apply.form.provider_access_dns01.label")
                  : fieldChallengeType === CHALLENGE_TYPE_HTTP01
                    ? t("workflow_node.apply.form.provider_access_http01.label")
                    : t("workflow_node.apply.form.provider_access.label")
              }
            >
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <AccessEditDrawer
                  mode="create"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.apply.form.provider_access.button")}
                      <IconPlus size="1.25em" />
                    </Button>
                  }
                  usage={fieldChallengeType === CHALLENGE_TYPE_DNS01 ? "dns" : fieldChallengeType === CHALLENGE_TYPE_HTTP01 ? "hosting" : "dns-hosting"}
                  afterSubmit={(record) => {
                    if (!accessOptionFilter(record.provider, record)) return;
                    if (fieldChallengeType === CHALLENGE_TYPE_DNS01 && acmeDns01ProvidersMap.get(fieldProvider!)?.provider !== record.provider) return;
                    if (fieldChallengeType === CHALLENGE_TYPE_HTTP01 && acmeHttp01ProvidersMap.get(fieldProvider!)?.provider !== record.provider) return;
                    formInst.setFieldValue("providerAccessId", record.id);
                  }}
                />
              </div>
              <Form.Item name="providerAccessId" dependencies={["challengeType", "provider"]} rules={[formRule]} noStyle>
                <AccessSelect
                  disabled={!fieldProvider}
                  placeholder={
                    fieldChallengeType === CHALLENGE_TYPE_DNS01
                      ? t("workflow_node.apply.form.provider_access_dns01.placeholder")
                      : fieldChallengeType === CHALLENGE_TYPE_HTTP01
                        ? t("workflow_node.apply.form.provider_access_http01.placeholder")
                        : t("workflow_node.apply.form.provider_access.placeholder")
                  }
                  showSearch
                  onFilter={accessOptionFilter}
                />
              </Form.Item>
            </Form.Item>

            <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
              {renderNestedFieldProviderComponent && <>{renderNestedFieldProviderComponent}</>}
            </FormNestedFieldsContextProvider>
          </div>

          <div id="certificate" data-anchor="certificate">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.certificate.title")}
              </Typography.Text>
            </Divider>

            <Form.Item name="keySource" label={t("workflow_node.apply.form.key_source.label")} rules={[formRule]}>
              <Radio.Group block onChange={(e) => handleKeySourceChange(e.target.value)}>
                <Radio.Button value={KEY_SOURCE_AUTO}>{t("workflow_node.apply.form.key_source.option.auto.label")}</Radio.Button>
                <Radio.Button value={KEY_SOURCE_REUSE}>{t("workflow_node.apply.form.key_source.option.reuse.label")}</Radio.Button>
                <Radio.Button value={KEY_SOURCE_CUSTOM}>{t("workflow_node.apply.form.key_source.option.custom.label")}</Radio.Button>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              name="keyAlgorithm"
              label={t("workflow_node.apply.form.key_algorithm.label")}
              extra={
                fieldKeySource === KEY_SOURCE_REUSE ? (
                  <span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.key_algorithm.help_reuse") }}></span>
                ) : fieldKeySource === KEY_SOURCE_CUSTOM ? (
                  <span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.key_algorithm.help_custom") }}></span>
                ) : (
                  void 0
                )
              }
              rules={[formRule]}
            >
              <Select
                options={["RSA2048", "RSA3072", "RSA4096", "RSA8192", "EC256", "EC384"].map((e) => ({
                  label: e,
                  value: e,
                }))}
                placeholder={t("workflow_node.apply.form.key_algorithm.placeholder")}
              />
            </Form.Item>

            <Show when={fieldKeySource === KEY_SOURCE_CUSTOM}>
              <Form.Item name="keyContent" label={t("workflow_node.apply.form.key_content.label")} rules={[formRule]}>
                <FileTextInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("workflow_node.apply.form.key_content.placeholder")} />
              </Form.Item>
            </Show>

            <Form.Item className="relative" label={t("workflow_node.apply.form.ca_provider.label")}>
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <Show when={!fieldCAProvider}>
                  <Link className="ant-typography" to="/settings/ssl-provider" target="_blank">
                    <Button size="small" type="link">
                      {t("workflow_node.apply.form.ca_provider.button")}
                      <IconChevronRight size="1.25em" />
                    </Button>
                  </Link>
                </Show>
              </div>
              <Form.Item name="caProvider" noStyle rules={[formRule]}>
                <CAProviderSelect
                  allowClear
                  placeholder={t("workflow_node.apply.form.ca_provider.placeholder")}
                  showAvailability
                  showDefault
                  showSearch
                  onSelect={handleCAProviderSelect}
                  onClear={handleCAProviderSelect}
                />
              </Form.Item>
            </Form.Item>

            <Form.Item label={t("workflow_node.apply.form.ca_provider_access.label")} hidden={!showCAProviderAccess}>
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <AccessEditDrawer
                  data={{ provider: caProvidersMap.get(fieldCAProvider!)?.provider }}
                  mode="create"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.apply.form.ca_provider_access.button")}
                      <IconChevronRight size="1.25em" />
                    </Button>
                  }
                  usage="ca"
                  afterSubmit={(record) => {
                    if (accessOptionFilterForCA(record.provider, record)) return;
                    if (caProvidersMap.get(fieldProvider!)?.provider !== record.provider) return;
                    formInst.setFieldValue("caProviderAccessId", record.id);
                  }}
                />
              </div>
              <Form.Item name="caProviderAccessId" dependencies={["caProvider"]} noStyle rules={[formRule]}>
                <AccessSelect
                  disabled={!fieldCAProvider}
                  placeholder={t("workflow_node.apply.form.ca_provider_access.placeholder")}
                  showSearch
                  onFilter={accessOptionFilterForCA}
                />
              </Form.Item>
            </Form.Item>

            <Form.Item
              name="validityLifetime"
              label={t("workflow_node.apply.form.validity_lifetime.label")}
              extra={t("workflow_node.apply.form.validity_lifetime.help")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.validity_lifetime.tooltip") }}></span>}
            >
              <InternalValidityLifetimeInput />
            </Form.Item>

            <Form.Item
              name="preferredChain"
              label={t("workflow_node.apply.form.preferred_chain.label")}
              extra={t("workflow_node.apply.form.preferred_chain.help")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.preferred_chain.tooltip") }}></span>}
            >
              <AutoComplete
                allowClear
                options={[
                  {
                    ca: "Let's Encrypt",
                    roots: ["ISRG", "ISRG Root X1", "ISRG Root X2"],
                  },
                  {
                    ca: "Google Trust Services",
                    roots: ["GTS", "GTS Root R1", "GTS Root R2", "GTS Root R3", "GTS Root R4", "GlobalSign", "GlobalSign R4"],
                  },
                ].map((e) => ({
                  label: e.ca,
                  options: e.roots.map((s) => ({
                    label: s,
                    value: s,
                  })),
                }))}
                placeholder={t("workflow_node.apply.form.preferred_chain.placeholder")}
                showSearch={{
                  filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
                }}
              />
            </Form.Item>

            <Form.Item
              name="acmeProfile"
              label={t("workflow_node.apply.form.acme_profile.label")}
              extra={t("workflow_node.apply.form.acme_profile.help")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.acme_profile.tooltip") }}></span>}
            >
              <AutoComplete
                allowClear
                options={[
                  {
                    ca: "Let's Encrypt",
                    profiles: ["classic", "tlsserver", "shortlived"],
                  },
                ].map((e) => ({
                  label: e.ca,
                  options: e.profiles.map((s) => ({
                    label: s,
                    value: s,
                  })),
                }))}
                placeholder={t("workflow_node.apply.form.acme_profile.placeholder")}
                showSearch={{
                  filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
                }}
              />
            </Form.Item>

            <Form.Item
              name="disableCommonName"
              label={t("workflow_node.apply.form.disable_cn.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.disable_cn.tooltip") }}></span>}
            >
              <Switch />
            </Form.Item>
          </div>

          <div id="advanced" data-anchor="advanced">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.advanced.title")}
              </Typography.Text>
            </Divider>

            <Form.Item
              name="nameservers"
              hidden={fieldIdentifier !== IDENTIFIER_DOMAIN}
              label={t("workflow_node.apply.form.nameservers.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.nameservers.tooltip") }}></span>}
            >
              <MultipleSplitValueInput
                modalTitle={t("workflow_node.apply.form.nameservers.multiple_input_modal.title")}
                placeholder={t("workflow_node.apply.form.nameservers.placeholder")}
                placeholderInModal={t("workflow_node.apply.form.nameservers.multiple_input_modal.placeholder")}
                separator={MULTIPLE_INPUT_SEPARATOR}
                splitOptions={{ removeEmpty: true, trimSpace: true }}
              />
            </Form.Item>

            <Form.Item
              name="dnsPropagationWait"
              hidden={fieldChallengeType !== CHALLENGE_TYPE_DNS01}
              label={t("workflow_node.apply.form.dns_propagation_wait.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.dns_propagation_wait.tooltip") }}></span>}
            >
              <Input
                type="number"
                allowClear
                min={0}
                max={3600}
                placeholder={t("workflow_node.apply.form.dns_propagation_wait.placeholder")}
                suffix={t("workflow_node.apply.form.dns_propagation_wait.unit")}
              />
            </Form.Item>

            <Form.Item
              name="dnsPropagationTimeout"
              hidden={fieldChallengeType !== CHALLENGE_TYPE_DNS01}
              label={t("workflow_node.apply.form.dns_propagation_timeout.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.dns_propagation_timeout.tooltip") }}></span>}
            >
              <Input
                type="number"
                allowClear
                min={0}
                max={3600}
                placeholder={t("workflow_node.apply.form.dns_propagation_timeout.placeholder")}
                suffix={t("workflow_node.apply.form.dns_propagation_timeout.unit")}
              />
            </Form.Item>

            <Form.Item
              name="dnsTTL"
              hidden={fieldChallengeType !== CHALLENGE_TYPE_DNS01}
              label={t("workflow_node.apply.form.dns_ttl.label")}
              extra={t("workflow_node.apply.form.dns_ttl.help")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.dns_ttl.tooltip") }}></span>}
            >
              <Input
                type="number"
                allowClear
                min={0}
                max={86400}
                placeholder={t("workflow_node.apply.form.dns_ttl.placeholder")}
                suffix={t("workflow_node.apply.form.dns_ttl.unit")}
              />
            </Form.Item>

            <Form.Item
              name="httpDelayWait"
              hidden={fieldChallengeType !== CHALLENGE_TYPE_HTTP01}
              label={t("workflow_node.apply.form.http_delay_wait.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.http_delay_wait.tooltip") }}></span>}
            >
              <Input
                type="number"
                allowClear
                min={0}
                max={3600}
                placeholder={t("workflow_node.apply.form.http_delay_wait.placeholder")}
                suffix={t("workflow_node.apply.form.http_delay_wait.unit")}
              />
            </Form.Item>

            <Form.Item
              name="disableFollowCNAME"
              hidden={fieldChallengeType !== CHALLENGE_TYPE_DNS01}
              label={t("workflow_node.apply.form.disable_follow_cname.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.disable_follow_cname.tooltip") }}></span>}
            >
              <Switch />
            </Form.Item>

            <Form.Item
              name="disableARI"
              label={t("workflow_node.apply.form.disable_ari.label")}
              rules={[formRule]}
              tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.disable_ari.tooltip") }}></span>}
            >
              <Switch />
            </Form.Item>
          </div>

          <div id="strategy" data-anchor="strategy">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.strategy.title")}
              </Typography.Text>
            </Divider>

            <Form.Item label={t("workflow_node.apply.form.skip_before_expiry_days.label")}>
              <span className="me-2 inline-block">{t("workflow_node.apply.form.skip_before_expiry_days.prefix")}</span>
              <span className="inline-block">
                <Form.Item name="skipBeforeExpiryDays" noStyle rules={[formRule]}>
                  <InputNumber
                    className="w-24"
                    min={1}
                    max={365}
                    placeholder={t("workflow_node.apply.form.skip_before_expiry_days.placeholder")}
                    suffix={t("workflow_node.apply.form.skip_before_expiry_days.unit")}
                  />
                </Form.Item>
              </span>
              <span className="ms-2 inline-block">{t("workflow_node.apply.form.skip_before_expiry_days.suffix")}</span>
            </Form.Item>
          </div>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const InternalIdentifierPicker = memo(({ disabled, onSelect }: { disabled?: boolean; onSelect?: (value: string) => void }) => {
  const { t } = useTranslation();

  const [value, setValue] = useState<string>();

  const options = [
    {
      value: IDENTIFIER_DOMAIN,
      label: t("workflow_node.apply.form.identifier.option.domain.label"),
      description: t("workflow_node.apply.form.identifier.option.domain.description"),
      icon: <IconWorldWww size="2rem" stroke="1.25" />,
    },
    {
      value: IDENTIFIER_IP,
      label: t("workflow_node.apply.form.identifier.option.ip.label"),
      description: t("workflow_node.apply.form.identifier.option.ip.description"),
      icon: <IconMapPin size="2rem" stroke="1.25" />,
    },
  ];

  const handleContinueClick = () => {
    if (!value) return;

    onSelect?.(value);
  };

  return (
    <>
      <Form.Item label={t("workflow_node.apply.form.identifier.label2")}>
        <div className="flex flex-col gap-2">
          {options.map((option) => (
            <Card
              className={mergeCls("relative overflow-hidden", { ["border-primary"]: value === option.value })}
              hoverable={!disabled}
              onClick={() => {
                if (disabled) return;

                setValue(option.value);
              }}
            >
              <div className="flex items-center gap-2">
                <div className="w-24 text-center">
                  <Avatar
                    style={{
                      background: "var(--color-primary)",
                    }}
                    icon={option.icon}
                    size={36}
                  />
                  <div className="mt-2 truncate text-sm font-medium">{option.label}</div>
                </div>
                <div className="flex-1 text-sm">
                  <Typography.Paragraph>
                    <blockquote dangerouslySetInnerHTML={{ __html: option.description }}></blockquote>
                  </Typography.Paragraph>
                </div>
              </div>
              {value === option.value && <div className="absolute top-0 left-0 size-full bg-primary opacity-20"></div>}
            </Card>
          ))}
        </div>
        <div className="mt-4 flex items-center justify-end gap-4">
          <Button disabled={!value || disabled} icon={<IconArrowRight size="1.25em" />} iconPlacement="end" type="primary" onClick={handleContinueClick}>
            {t("workflow_node.apply.form.identifier.continue.button")}
          </Button>
        </div>
      </Form.Item>
    </>
  );
});

const InternalEmailInput = memo(({ disabled, ...props }: { disabled?: boolean; value?: string; onChange?: (value: string) => void }) => {
  const { t } = useTranslation();

  const { emails, fetchEmails, removeEmail } = useContactEmailsStore();
  useMount(() => {
    fetchEmails(false);
  });

  const [value, setValue] = useControllableValue<string>(props, {
    valuePropName: "value",
    defaultValuePropName: "defaultValue",
    trigger: "onChange",
  });

  const [inputValue, setInputValue] = useState<string>();

  const renderOptionLabel = (email: string, removable: boolean = false) => (
    <div className="flex items-center gap-2 overflow-hidden">
      <span className="flex-1 truncate overflow-hidden">{email}</span>
      {removable && (
        <Button
          color="default"
          disabled={disabled}
          icon={<IconCircleMinus size="1.25em" />}
          size="small"
          type="text"
          onClick={(e) => {
            removeEmail(email);
            e.stopPropagation();
          }}
        />
      )}
    </div>
  );

  const options = useMemo(() => {
    const temp = emails.map((email) => ({
      label: renderOptionLabel(email, true),
      value: email,
    }));

    if (!!inputValue && temp.every((option) => option.value !== inputValue)) {
      temp.unshift({
        label: renderOptionLabel(inputValue),
        value: inputValue,
      });
    }

    return temp;
  }, [emails, inputValue]);

  const handleChange = (value: string) => {
    setValue(value);
  };

  const handleSearch = (value: string) => {
    setInputValue(value?.trim());
  };

  return (
    <AutoComplete
      backfill
      defaultValue={value}
      disabled={disabled}
      options={options}
      placeholder={t("workflow_node.apply.form.contact_email.placeholder")}
      showSearch={{
        filterOption: true,
        onSearch: handleSearch,
      }}
      value={value}
      onChange={handleChange}
    />
  );
});

const InternalValidityLifetimeInput = memo(({ disabled, ...props }: { disabled?: boolean; value?: string; onChange?: (value: string) => void }) => {
  const { t } = useTranslation();

  const [value, setValue] = useControllableValue<string>(props, {
    valuePropName: "value",
    defaultValuePropName: "defaultValue",
    trigger: "onChange",
  });

  const parseCombinedValue = (val: string): [string | undefined, string | undefined] => {
    const match = String(val).match(/^(\d+)([a-zA-Z]+)$/);
    if (match) {
      return [match[1], match[2]];
    }

    return [undefined, undefined];
  };

  const [inputValue, setInputValue] = useState(parseCombinedValue(value)[0]);
  const [selectValue, setSelectValue] = useState(parseCombinedValue(value)[1] || "d");
  useEffect(() => {
    const [v, u] = parseCombinedValue(value);
    setInputValue(v);
    setSelectValue(u || "d");
  }, [value]);

  const handleInputClear = () => {
    setValue("");
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.currentTarget.value);

    if (e.currentTarget.value) {
      setValue(`${e.currentTarget.value}${selectValue}`);
    } else {
      setValue("");
    }
  };

  const handleSelectChange = (value: string) => {
    setSelectValue(value);

    if (inputValue) {
      setValue(`${inputValue}${value}`);
    }
  };

  return (
    <Space.Compact className="w-full">
      <Input
        allowClear
        disabled={disabled}
        placeholder={t("workflow_node.apply.form.validity_lifetime.placeholder")}
        type="number"
        value={inputValue}
        onChange={handleInputChange}
        onClear={handleInputClear}
      />
      <div className="w-24">
        <Select
          options={["h", "d"].map((s) => ({
            key: s,
            label: t(`workflow_node.apply.form.validity_lifetime.units.${s}`),
            value: s,
          }))}
          value={selectValue}
          onChange={handleSelectChange}
        />
      </div>
    </Space.Compact>
  );
});

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters", "challenge", "certificate", "advanced", "strategy"].map((key) => ({
    key: key,
    title: t(`workflow_node.apply.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    contactEmail: "",
    ...(defaultNodeConfigForBizApply() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      identifier: z.enum([IDENTIFIER_DOMAIN, IDENTIFIER_IP]),
      domains: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isDomain(e, { allowWildcard: true }));
        }, t("common.errmsg.domain_invalid")),
      ipaddrs: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isIPv4(e) || isIPv6(e));
        }, t("common.errmsg.ip_invalid")),
      contactEmail: z.email(t("common.errmsg.email_invalid")),
      challengeType: z.enum([CHALLENGE_TYPE_DNS01, CHALLENGE_TYPE_HTTP01], t("workflow_node.apply.form.challenge_type.placeholder")),
      provider: z.string().nonempty(t("workflow_node.apply.form.provider.placeholder")),
      providerAccessId: z.string().nullish(),
      providerConfig: z.any().nullish(),
      caProvider: z.string().nullish(),
      caProviderAccessId: z.string().nullish(),
      caProviderConfig: z.any().nullish(),
      keySource: z.enum([KEY_SOURCE_AUTO, KEY_SOURCE_REUSE, KEY_SOURCE_CUSTOM], t("workflow_node.apply.form.key_source.placeholder")),
      keyAlgorithm: z.string().nonempty(t("workflow_node.apply.form.key_algorithm.placeholder")),
      keyContent: z.string().nullish(),
      validityLifetime: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return /^\d+[d|h]$/.test(v) && parseInt(v) > 0;
        }, t("workflow_node.apply.form.validity_lifetime.placeholder")),
      preferredChain: z.string().nullish(),
      acmeProfile: z.string().nullish(),
      nameservers: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isHostname(e) || isDomain(e));
        }, t("common.errmsg.host_invalid")),
      dnsPropagationWait: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z.number().int().gte(0, t("workflow_node.apply.form.dns_propagation_wait.placeholder")).nullish()
      ),
      dnsPropagationTimeout: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z.number().int().gte(1, t("workflow_node.apply.form.dns_propagation_timeout.placeholder")).nullish()
      ),
      dnsTTL: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z.number().int().gte(1, t("workflow_node.apply.form.dns_ttl.placeholder")).nullish()
      ),
      httpDelayWait: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z.number().int().gte(0, t("workflow_node.apply.form.http_delay_wait.placeholder")).nullish()
      ),
      disableCommonName: z.boolean().nullish(),
      disableFollowCNAME: z.boolean().nullish(),
      disableARI: z.boolean().nullish(),
      skipBeforeExpiryDays: z.coerce.number().int().positive(),
    })
    .superRefine((values, ctx) => {
      if (values.identifier) {
        switch (values.identifier) {
          case IDENTIFIER_DOMAIN:
            {
              if (!values.domains) {
                ctx.addIssue({
                  code: "custom",
                  message: t("common.errmsg.domain_invalid"),
                  path: ["domains"],
                });
              }
            }
            break;
          case IDENTIFIER_IP:
            {
              if (!values.ipaddrs) {
                ctx.addIssue({
                  code: "custom",
                  message: t("common.errmsg.ip_invalid"),
                  path: ["ipaddrs"],
                });
              }
            }
            break;
        }
      }

      if (values.challengeType) {
        switch (values.challengeType) {
          case CHALLENGE_TYPE_DNS01:
            {
              if (values.ipaddrs) {
                ctx.addIssue({
                  code: "custom",
                  message: t("workflow_node.apply.form.challenge_type.errmsg.no_ip_in_dns01"),
                  path: ["challengeType"],
                });
              }
            }
            break;

          case CHALLENGE_TYPE_HTTP01:
            {
              if (values.domains && values.domains.includes("*")) {
                ctx.addIssue({
                  code: "custom",
                  message: t("workflow_node.apply.form.challenge_type.errmsg.no_wildcard_in_http01"),
                  path: ["challengeType"],
                });
              }
            }
            break;
        }
      }

      if (values.keySource) {
        switch (values.keySource) {
          case KEY_SOURCE_CUSTOM:
            {
              if (!validatePEMPrivateKey(values.keyContent!)) {
                ctx.addIssue({
                  code: "custom",
                  message: t("workflow_node.apply.form.key_content.errmsg.invalid"),
                  path: ["keyContent"],
                });
              } else {
                const { algorithm, keySize } = getPKIXPrivateKeyAlgorithm(values.keyContent!);
                const expectedKeyAlg = values.keyAlgorithm;
                const actualKeyAlg = `${algorithm}${keySize}`;
                if (actualKeyAlg !== expectedKeyAlg) {
                  ctx.addIssue({
                    code: "custom",
                    message: t("workflow_node.apply.form.key_content.errmsg.not_matched", { expected: expectedKeyAlg, actual: actualKeyAlg }),
                    path: ["keyContent"],
                  });
                }
              }
            }
            break;
        }
      }

      if (values.provider) {
        switch (values.challengeType) {
          case CHALLENGE_TYPE_DNS01:
            {
              const provider = acmeDns01ProvidersMap.get(values.provider);
              if (!provider?.builtin && !values.providerAccessId) {
                ctx.addIssue({
                  code: "custom",
                  message: t("workflow_node.deploy.form.provider_access.placeholder"),
                  path: ["providerAccessId"],
                });
              }
            }
            break;

          case CHALLENGE_TYPE_HTTP01:
            {
              const provider = acmeHttp01ProvidersMap.get(values.provider);
              if (!provider?.builtin && !values.providerAccessId) {
                ctx.addIssue({
                  code: "custom",
                  message: t("workflow_node.deploy.form.provider_access.placeholder"),
                  path: ["providerAccessId"],
                });
              }
            }
            break;
        }
      }

      if (values.caProvider) {
        const provider = caProvidersMap.get(values.caProvider);
        if (!provider?.builtin && !values.caProviderAccessId) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.apply.form.ca_provider_access.placeholder"),
            path: ["caProviderAccessId"],
          });
        }
      }
    });
};

const _default = Object.assign(BizApplyNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
