import { memo, useEffect, useMemo, useRef, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { Link } from "react-router";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconArrowRight, IconChevronRight, IconCircleMinus, IconMapPin, IconPlus, IconWorldWww } from "@tabler/icons-react";
import { useControllableValue, useMount } from "ahooks";
import {
  Alert,
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
import { ACCESS_PROVIDERS, CA_PROVIDERS, acmeDns01ProvidersMap, acmeHttp01ProvidersMap, caProvidersMap } from "@/domain/provider";
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

const SOURCE_ACME = "acme" as const;
const SOURCE_TENCENTCLOUD = "tencentcloud" as const;
const TENCENTCLOUD_KEY_ALGORITHMS = ["RSA2048", "EC256"] as const;
const TENCENTCLOUD_DNS_PROVIDERS = ["tencentcloud-dns", "aliyun-dns", "cloudflare"] as const;
const isTencentCloudDnsProvider = (provider?: string) => TENCENTCLOUD_DNS_PROVIDERS.some((value) => value === provider);

const accessOptionFilterForTencentCloud = (_: string, option: AccessModel) => !option.reserve && option.provider === ACCESS_PROVIDERS.TENCENTCLOUD;

const IDENTIFIER_DOMAIN = "domain" as const;
const IDENTIFIER_IP = "ip" as const;

const CHALLENGE_TYPE_DNS_AUTO = "dns-auto" as const;
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
    if (isTencentCloud && (fieldChallengeType !== CHALLENGE_TYPE_DNS01 || !isTencentCloudDnsProvider(fieldProvider))) return false;
    if (fieldChallengeType === CHALLENGE_TYPE_DNS01) return acmeDns01ProvidersMap.get(fieldProvider)?.provider === option.provider;
    if (fieldChallengeType === CHALLENGE_TYPE_HTTP01) return acmeHttp01ProvidersMap.get(fieldProvider)?.provider === option.provider;
    return false;
  };
  const accessOptionFilterForCA = (_: string, option: AccessModel) => {
    if (option.reserve !== "ca") return false;
    return caProvidersMap.get(fieldCAProvider!)?.provider === option.provider;
  };

  const initialValues = useMemo(() => {
    const config = node.form?.getValueIn("config") as WorkflowNodeConfigForBizApply | undefined;
    if (config?.source === SOURCE_TENCENTCLOUD) {
      const manualDNS = config.challengeType === CHALLENGE_TYPE_DNS01;
      const supportedProvider = manualDNS && isTencentCloudDnsProvider(config.provider);
      return {
        ...config,
        identifier: IDENTIFIER_DOMAIN,
        challengeType: manualDNS ? CHALLENGE_TYPE_DNS01 : CHALLENGE_TYPE_DNS_AUTO,
        provider: supportedProvider ? config.provider : void 0,
        providerAccessId: supportedProvider ? config.providerAccessId : void 0,
        providerConfig: void 0,
        nameservers: void 0,
        dnsPropagationTimeout: void 0,
        httpDelayWait: void 0,
        disableFollowCNAME: void 0,
        dnsPropagationWait: manualDNS ? config.dnsPropagationWait : void 0,
        dnsTTL: manualDNS ? config.dnsTTL : void 0,
        keySource: KEY_SOURCE_AUTO,
        keyAlgorithm: config.keyAlgorithm === "EC256" ? "EC256" : "RSA2048",
      };
    }
    return config;
  }, [node]);
  const previousSource = useRef(initialValues?.source ?? SOURCE_ACME);

  const { form: formInst, formProps } = useAntdForm<WorkflowNodeConfigForBizApply>({
    form: props.form,
    name: "workflowNodeBizApplyConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldSource = Form.useWatch("source", { form: formInst, preserve: true });
  const isTencentCloud = (fieldSource ?? initialValues?.source) === SOURCE_TENCENTCLOUD;
  // antd-zod requires an object schema; workflow validation uses the discriminated union below.
  const formRule = isTencentCloud ? createSchemaFieldRule(getTencentCloudSchema({ i18n })) : createSchemaFieldRule(getACMESchema({ i18n }));

  const fieldIdentifier = Form.useWatch("identifier", { form: formInst, preserve: true });
  const fieldChallengeType = Form.useWatch<typeof CHALLENGE_TYPE_DNS_AUTO | typeof CHALLENGE_TYPE_DNS01 | typeof CHALLENGE_TYPE_HTTP01>("challengeType", {
    form: formInst,
    preserve: true,
  });
  const fieldProvider = Form.useWatch("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch("providerAccessId", { form: formInst, preserve: true });
  const fieldKeySource = Form.useWatch("keySource", { form: formInst, preserve: true });
  const fieldCAProvider = Form.useWatch("caProvider", { form: formInst, preserve: true });
  const fieldCAProviderAccessId = Form.useWatch("caProviderAccessId", { form: formInst, preserve: true });

  const renderNestedFieldProviderComponent = BizApplyNodeConfigFieldsProvider.useComponent(
    fieldChallengeType === CHALLENGE_TYPE_HTTP01 ? CHALLENGE_TYPE_HTTP01 : CHALLENGE_TYPE_DNS01,
    isTencentCloud ? "" : fieldProvider,
    {}
  );

  const resetFieldIfInvalid = (field: keyof z.infer<ReturnType<typeof getACMESchema>>) => {
    const fieldSchema = getACMESchema({ i18n }).shape[field];
    const fieldValue = formInst.getFieldValue(field);
    if (!fieldSchema.safeParse(fieldValue).success) {
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

  const fieldCAProviderAccessOptional = useMemo(() => {
    if (fieldCAProvider) {
      const provider = caProvidersMap.get(fieldCAProvider);
      return provider?.accessOptional ?? false;
    }

    return false;
  }, [fieldCAProvider]);

  useEffect(() => {
    if (isTencentCloud && fieldChallengeType !== CHALLENGE_TYPE_DNS01) return;

    // 如果未选择质询提供商，则清空授权信息
    if (!fieldProvider && fieldProviderAccessId) {
      formInst.setFieldValue("providerAccessId", void 0);
      return;
    }

    // 原生 DNS 授权必须与当前 DNS 提供商匹配，与 SSL 授权相互独立。
    const selectedAccess = accesses.find((access) => access.id === fieldProviderAccessId);
    if (isTencentCloud && selectedAccess && !accessOptionFilter(selectedAccess.provider, selectedAccess)) {
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
  }, [isTencentCloud, fieldChallengeType, fieldProvider, fieldProviderAccessId, accesses]);

  useEffect(() => {
    if (isTencentCloud) return;

    // 如果未选择 CA 提供商，则清空授权信息
    if (!fieldCAProvider && fieldCAProviderAccessId) {
      formInst.setFieldValue("caProviderAccessId", void 0);
      return;
    }

    // 如果已选择 CA 提供商只有一个授权信息，则自动选择该授权信息
    if (fieldCAProvider && !fieldCAProviderAccessId && !fieldCAProviderAccessOptional) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilterForCA(access.provider, access))
        .filter((access) => caProvidersMap.get(fieldCAProvider)?.provider === access.provider);
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("caProviderAccessId", availableAccesses[0].id);
      }
    }
  }, [isTencentCloud, fieldCAProvider, fieldCAProviderAccessId, fieldCAProviderAccessOptional]);

  useEffect(() => {
    if (!fieldSource) return;
    const sourceChanged = previousSource.current !== fieldSource;
    previousSource.current = fieldSource;

    if (fieldSource === SOURCE_TENCENTCLOUD) {
      // 仅切换来源时重置为 DNS_AUTO，重新打开已保存的自选 DNS 配置时保留原模式。
      const manualDNS = !sourceChanged && formInst.getFieldValue("challengeType") === CHALLENGE_TYPE_DNS01;
      formInst.setFieldsValue({
        identifier: IDENTIFIER_DOMAIN,
        ipaddrs: void 0,
        challengeType: manualDNS ? CHALLENGE_TYPE_DNS01 : CHALLENGE_TYPE_DNS_AUTO,
        providerConfig: void 0,
        nameservers: void 0,
        dnsPropagationTimeout: void 0,
        httpDelayWait: void 0,
        disableFollowCNAME: void 0,
        keySource: KEY_SOURCE_AUTO,
        keyContent: void 0,
        keyAlgorithm: formInst.getFieldValue("keyAlgorithm") === "EC256" ? "EC256" : "RSA2048",
      });
      if (!manualDNS || !isTencentCloudDnsProvider(formInst.getFieldValue("provider"))) {
        formInst.setFieldsValue({ provider: void 0, providerAccessId: void 0 });
      }
      if (!manualDNS) {
        formInst.setFieldsValue({ dnsPropagationWait: void 0, dnsTTL: void 0 });
      }
    } else if (fieldSource === SOURCE_ACME) {
      const challengeType = formInst.getFieldValue("challengeType");
      if (!challengeType || challengeType === CHALLENGE_TYPE_DNS_AUTO) {
        formInst.setFieldValue("challengeType", CHALLENGE_TYPE_DNS01);
      }
      if (sourceChanged) {
        formInst.setFieldsValue({ provider: void 0, providerAccessId: void 0, providerConfig: void 0 });
      }
    }
  }, [fieldSource, fieldChallengeType, formInst]);

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
    if (isTencentCloud) {
      formInst.setFieldsValue({ provider: void 0, providerAccessId: void 0, providerConfig: void 0 });
      return;
    }

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
    // 切换质询提供商时重置表单，避免其他提供商的配置字段残留
    if (!isTencentCloud && initialValues?.provider === value) {
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
        <Form.Item
          name="source"
          label={t("workflow_node.apply.form.source.label")}
          getValueProps={(value) => ({ value: value ?? SOURCE_ACME })}
          rules={[createSchemaFieldRule(z.object({ source: z.enum([SOURCE_ACME, SOURCE_TENCENTCLOUD]).optional() }))]}
        >
          <Radio.Group block>
            <Radio.Button value={SOURCE_ACME}>{t("workflow_node.apply.form.source.option.acme")}</Radio.Button>
            <Radio.Button value={SOURCE_TENCENTCLOUD}>{t("workflow_node.apply.form.source.option.tencentcloud")}</Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Show when={!isTencentCloud && !fieldIdentifier}>
          <InternalIdentifierPicker onSelect={handleIdentifierPick} />
        </Show>

        <Show when={isTencentCloud}>
          <Alert
            className="mb-6"
            type="info"
            showIcon
            title={t("workflow_node.apply.form.tencentcloud_note.title")}
            description={
              <div className="space-y-2">
                <p>{t("workflow_node.apply.form.tencentcloud_note.dns")}</p>
                <p>{t("workflow_node.apply.form.tencentcloud_note.eligibility")}</p>
                <p>{t("workflow_node.apply.form.tencentcloud_note.limits")}</p>
                <p>{t("workflow_node.apply.form.tencentcloud_note.issuance")}</p>
                <p>{t("workflow_node.apply.form.tencentcloud_note.permissions")}</p>
              </div>
            }
          />
        </Show>

        <div style={{ display: isTencentCloud || fieldIdentifier ? "block" : "none" }}>
          <div id="parameters" data-anchor="parameters">
            <Form.Item name="identifier" hidden={!isTencentCloud} label={t("workflow_node.apply.form.identifier.label")} rules={[formRule]}>
              <Radio.Group block disabled={isTencentCloud} onChange={(e) => handleIdentifierChange(e.target.value)}>
                <Radio.Button value={IDENTIFIER_DOMAIN}>{t("workflow_node.apply.form.identifier.option.domain.label")}</Radio.Button>
                {!isTencentCloud && <Radio.Button value={IDENTIFIER_IP}>{t("workflow_node.apply.form.identifier.option.ip.label")}</Radio.Button>}
              </Radio.Group>
            </Form.Item>

            <Show>
              <Show.Case when={fieldIdentifier === IDENTIFIER_DOMAIN}>
                <Form.Item
                  name="domains"
                  dependencies={["source", "identifier", "challengeType"]}
                  label={t("workflow_node.apply.form.domains.label")}
                  extra={
                    isTencentCloud ? (
                      t("workflow_node.apply.form.tencentcloud_domain.help")
                    ) : (
                      <span
                        dangerouslySetInnerHTML={{
                          __html:
                            fieldChallengeType === CHALLENGE_TYPE_HTTP01
                              ? t("workflow_node.apply.form.domains.help_no_wildcard")
                              : t("workflow_node.apply.form.domains.help"),
                        }}
                      ></span>
                    )
                  }
                  rules={[formRule]}
                >
                  {isTencentCloud ? (
                    <Input placeholder={t("workflow_node.apply.form.tencentcloud_domain.placeholder")} />
                  ) : (
                    <MultipleSplitValueInput
                      modalTitle={t("workflow_node.apply.form.domains.multiple_input_modal.title")}
                      placeholder={t("workflow_node.apply.form.domains.placeholder")}
                      placeholderInModal={t("workflow_node.apply.form.domains.multiple_input_modal.placeholder")}
                      separator={MULTIPLE_INPUT_SEPARATOR}
                      splitOptions={{ removeEmpty: true, trimSpace: true }}
                    />
                  )}
                </Form.Item>
              </Show.Case>
              <Show.Case when={!isTencentCloud && fieldIdentifier === IDENTIFIER_IP}>
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
              tooltip={isTencentCloud ? void 0 : <span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.contact_email.tooltip") }}></span>}
            >
              <InternalEmailInput />
            </Form.Item>

            <Show when={isTencentCloud}>
              <Form.Item className="relative" label={t("workflow_node.apply.form.tencentcloud_access.label")}>
                <div className="absolute -top-1.5 right-0 -translate-y-full">
                  <AccessEditDrawer
                    data={{ provider: ACCESS_PROVIDERS.TENCENTCLOUD }}
                    mode="create"
                    usage="dns"
                    trigger={
                      <Button size="small" type="link">
                        {t("workflow_node.apply.form.provider_access.button")}
                        <IconPlus size="1.25em" />
                      </Button>
                    }
                    afterSubmit={(record) => {
                      if (accessOptionFilterForTencentCloud(record.provider, record)) {
                        formInst.setFieldValue("tencentcloudAccessId", record.id);
                      }
                    }}
                  />
                </div>
                <Form.Item name="tencentcloudAccessId" rules={[formRule]} noStyle>
                  <AccessSelect
                    placeholder={t("workflow_node.apply.form.tencentcloud_access.placeholder")}
                    showSearch
                    onFilter={accessOptionFilterForTencentCloud}
                  />
                </Form.Item>
              </Form.Item>
            </Show>
          </div>

          <div id="challenge" data-anchor="challenge">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.challenge.title")}
              </Typography.Text>
            </Divider>

            <Form.Item
              name="challengeType"
              initialValue={isTencentCloud ? CHALLENGE_TYPE_DNS_AUTO : CHALLENGE_TYPE_DNS01}
              dependencies={["source", "identifier", "domains", "ipaddrs"]}
              label={t("workflow_node.apply.form.challenge_type.label")}
              extra={isTencentCloud ? t("workflow_node.apply.form.tencentcloud_challenge.help") : void 0}
              rules={[formRule]}
              tooltip={isTencentCloud ? void 0 : <span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.challenge_type.tooltip") }}></span>}
            >
              <Radio.Group block onChange={(e) => handleChallengeTypeChange(e.target.value)}>
                {isTencentCloud ? (
                  <>
                    <Radio.Button value={CHALLENGE_TYPE_DNS_AUTO}>{t("workflow_node.apply.form.tencentcloud_challenge.option.auto")}</Radio.Button>
                    <Radio.Button value={CHALLENGE_TYPE_DNS01}>{t("workflow_node.apply.form.tencentcloud_challenge.option.manual")}</Radio.Button>
                  </>
                ) : (
                  <>
                    <Radio.Button disabled={fieldIdentifier === IDENTIFIER_IP} value={CHALLENGE_TYPE_DNS01}>
                      DNS-01
                    </Radio.Button>
                    <Radio.Button value={CHALLENGE_TYPE_HTTP01}>HTTP-01</Radio.Button>
                  </>
                )}
              </Radio.Group>
            </Form.Item>

            <Show when={!isTencentCloud || fieldChallengeType === CHALLENGE_TYPE_DNS01}>
              <Form.Item
                name="provider"
                dependencies={["source", "challengeType"]}
                label={
                  fieldChallengeType === CHALLENGE_TYPE_DNS01
                    ? t("workflow_node.apply.form.provider_dns01.label")
                    : fieldChallengeType === CHALLENGE_TYPE_HTTP01
                      ? t("workflow_node.apply.form.provider_http01.label")
                      : t("workflow_node.apply.form.provider.label")
                }
                required={isTencentCloud}
                rules={[formRule]}
              >
                {fieldChallengeType === CHALLENGE_TYPE_DNS01 ? (
                  <ACMEDns01ProviderSelect
                    placeholder={t("workflow_node.apply.form.provider_dns01.placeholder")}
                    showAvailability
                    showSearch
                    onFilter={isTencentCloud ? (value) => isTencentCloudDnsProvider(value) : void 0}
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
                extra={isTencentCloud ? t("workflow_node.apply.form.tencentcloud_dns_access.help") : void 0}
              >
                <div className="absolute -top-1.5 right-0 -translate-y-full">
                  <AccessEditDrawer
                    data={isTencentCloud ? { provider: acmeDns01ProvidersMap.get(fieldProvider)?.provider } : void 0}
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
                <Form.Item name="providerAccessId" dependencies={["source", "challengeType", "provider"]} rules={[formRule]} noStyle>
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

              <Show when={!isTencentCloud}>
                <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
                  {renderNestedFieldProviderComponent && <>{renderNestedFieldProviderComponent}</>}
                </FormNestedFieldsContextProvider>
              </Show>
            </Show>

            <Show when={isTencentCloud && fieldChallengeType === CHALLENGE_TYPE_DNS01}>
              <Form.Item
                name="dnsPropagationWait"
                label={t("workflow_node.apply.form.dns_propagation_wait.label")}
                extra={t("workflow_node.apply.form.tencentcloud_dns_propagation_wait.help")}
                rules={[formRule]}
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
                name="dnsTTL"
                label={t("workflow_node.apply.form.dns_ttl.label")}
                extra={t("workflow_node.apply.form.tencentcloud_dns_ttl.help")}
                rules={[formRule]}
              >
                <Input type="number" allowClear min={1} max={86400} placeholder="600" suffix={t("workflow_node.apply.form.dns_ttl.unit")} />
              </Form.Item>
            </Show>
          </div>

          <div id="certificate" data-anchor="certificate">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.certificate.title")}
              </Typography.Text>
            </Divider>

            <Form.Item name="keySource" hidden={isTencentCloud} label={t("workflow_node.apply.form.key_source.label")} rules={[formRule]}>
              {isTencentCloud ? (
                <Input readOnly />
              ) : (
                <Radio.Group block onChange={(e) => handleKeySourceChange(e.target.value)}>
                  <Radio.Button value={KEY_SOURCE_AUTO}>{t("workflow_node.apply.form.key_source.option.auto.label")}</Radio.Button>
                  <Radio.Button value={KEY_SOURCE_REUSE}>{t("workflow_node.apply.form.key_source.option.reuse.label")}</Radio.Button>
                  <Radio.Button value={KEY_SOURCE_CUSTOM}>{t("workflow_node.apply.form.key_source.option.custom.label")}</Radio.Button>
                </Radio.Group>
              )}
            </Form.Item>

            <Form.Item
              name="keyAlgorithm"
              dependencies={["source"]}
              label={t("workflow_node.apply.form.key_algorithm.label")}
              extra={
                isTencentCloud ? (
                  t("workflow_node.apply.form.tencentcloud_key.help")
                ) : fieldKeySource === KEY_SOURCE_REUSE ? (
                  <span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.key_source.option.reuse.help") }}></span>
                ) : fieldKeySource === KEY_SOURCE_CUSTOM ? (
                  <span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.key_source.option.custom.help") }}></span>
                ) : (
                  void 0
                )
              }
              rules={[formRule]}
            >
              <Select
                options={(isTencentCloud ? TENCENTCLOUD_KEY_ALGORITHMS : ["RSA2048", "RSA3072", "RSA4096", "RSA8192", "EC256", "EC384"]).map((e) => ({
                  label: e,
                  value: e,
                }))}
                placeholder={t("workflow_node.apply.form.key_algorithm.placeholder")}
              />
            </Form.Item>

            <Show when={!isTencentCloud && fieldKeySource === KEY_SOURCE_CUSTOM}>
              <Form.Item name="keyContent" label={t("workflow_node.apply.form.key_content.label")} rules={[formRule]}>
                <FileTextInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("workflow_node.apply.form.key_content.placeholder")} />
              </Form.Item>
            </Show>

            <Show when={!isTencentCloud}>
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

              <Form.Item
                label={t("workflow_node.apply.form.ca_provider_access.label")}
                hidden={!showCAProviderAccess}
                extra={
                  fieldCAProviderAccessOptional
                    ? t("access.form.shared_ca_access_optional.help", {
                        provider: t(caProvidersMap.get(fieldCAProvider!)?.name ?? ""),
                      })
                    : void 0
                }
              >
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
                    allowClear={fieldCAProviderAccessOptional}
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
            </Show>
          </div>

          <Show when={!isTencentCloud}>
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
          </Show>

          <div id="strategy" data-anchor="strategy">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.apply.form_anchor.strategy.title")}
              </Typography.Text>
            </Divider>

            <Form.Item label={t("workflow_node.apply.form.skip_before_expiry_days.label")} extra={t("workflow_node.apply.form.skip_before_expiry_days.help")}>
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
              key={option.value}
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

const getAnchorItems = ({
  i18n = getI18n(),
  source,
}: {
  i18n?: ReturnType<typeof getI18n>;
  source?: WorkflowNodeConfigForBizApply["source"];
}): Required<AnchorProps>["items"] => {
  const { t } = i18n;
  const keys =
    source === SOURCE_TENCENTCLOUD
      ? ["parameters", "challenge", "certificate", "strategy"]
      : ["parameters", "challenge", "certificate", "advanced", "strategy"];

  return keys.map((key) => ({
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
  return z.discriminatedUnion("source", [getACMESchema({ i18n }), getTencentCloudSchema({ i18n })]);
};

const getTencentCloudSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      source: z.literal(SOURCE_TENCENTCLOUD),
      identifier: z.literal(IDENTIFIER_DOMAIN),
      domains: z.string().refine((value) => {
        return (
          isDomain(value) &&
          value.includes(".") &&
          value.length <= 253 &&
          value.split(".").every((label) => label.length <= 63 && !label.startsWith("-")) &&
          !isIPv4(value) &&
          !isIPv6(value)
        );
      }, t("workflow_node.apply.form.tencentcloud_domain.errmsg")),
      ipaddrs: z.literal("").nullish(),
      contactEmail: z.email(),
      tencentcloudAccessId: z
        .string({ error: t("workflow_node.apply.form.tencentcloud_access.placeholder") })
        .trim()
        .nonempty(t("workflow_node.apply.form.tencentcloud_access.placeholder")),
      challengeType: z.preprocess((value) => (value === "" ? void 0 : value), z.enum([CHALLENGE_TYPE_DNS_AUTO, CHALLENGE_TYPE_DNS01]).optional()),
      provider: z.enum(TENCENTCLOUD_DNS_PROVIDERS, { error: t("workflow_node.apply.form.tencentcloud_dns_provider.errmsg") }).nullish(),
      providerAccessId: z.string().nullish(),
      dnsPropagationWait: z.coerce.number().int().min(0).max(3600).or(z.literal("")).nullish(),
      dnsTTL: z.coerce.number().int().min(1).max(86400).or(z.literal("")).nullish(),
      keySource: z.literal(KEY_SOURCE_AUTO),
      keyAlgorithm: z.enum(TENCENTCLOUD_KEY_ALGORITHMS),
      keyContent: z.literal("").nullish(),
      skipBeforeExpiryDays: z.coerce.number().int().positive(),
    })
    .superRefine((values, ctx) => {
      if (values.challengeType !== CHALLENGE_TYPE_DNS01) return;
      if (!values.provider) {
        ctx.addIssue({
          code: "custom",
          message: t("workflow_node.apply.form.tencentcloud_dns_provider.errmsg"),
          path: ["provider"],
        });
      }
      if (!values.providerAccessId?.trim()) {
        ctx.addIssue({
          code: "custom",
          message: t("workflow_node.apply.form.provider_access_dns01.placeholder"),
          path: ["providerAccessId"],
        });
      }
    });
};

const getACMESchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      source: z.literal(SOURCE_ACME).optional(),
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
      contactEmail: z.email(),
      challengeType: z.enum([CHALLENGE_TYPE_DNS01, CHALLENGE_TYPE_HTTP01]),
      provider: z.string().nonempty(),
      providerAccessId: z.string().nullish(),
      providerConfig: z.any().nullish(),
      caProvider: z.string().nullish(),
      caProviderAccessId: z.string().nullish(),
      caProviderConfig: z.any().nullish(),
      keySource: z.enum([KEY_SOURCE_AUTO, KEY_SOURCE_REUSE, KEY_SOURCE_CUSTOM]),
      keyAlgorithm: z.string().nonempty(),
      keyContent: z.string().nullish(),
      validityLifetime: z
        .templateLiteral([z.nullish(z.number().positive()), z.enum(["h", "d"])])
        .or(z.literal(""))
        .nullish(),
      preferredChain: z.string().nullish(),
      acmeProfile: z.string().nullish(),
      nameservers: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isHostname(e) || isDomain(e));
        }, t("common.errmsg.host_invalid")),
      dnsPropagationWait: z.coerce.number().int().gte(0).or(z.literal("")).nullish(),
      dnsPropagationTimeout: z.coerce.number().int().gt(0).or(z.literal("")).nullish(),
      dnsTTL: z.coerce.number().int().gt(0).or(z.literal("")).nullish(),
      httpDelayWait: z.coerce.number().int().gt(0).or(z.literal("")).nullish(),
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
              const scDomains = z.string().nonempty();
              const spDomains = scDomains.safeParse(values.domains);
              if (!spDomains.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spDomains.error).errors.join(),
                  path: ["domains"],
                });
              }
            }
            break;

          case IDENTIFIER_IP:
            {
              const scIpAddrs = z.string().nonempty();
              const spIpAddrs = scIpAddrs.safeParse(values.ipaddrs);
              if (!spIpAddrs.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spIpAddrs.error).errors.join(),
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
        if (!provider?.builtin && !provider?.accessOptional && !values.caProviderAccessId) {
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
