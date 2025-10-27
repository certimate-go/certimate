import { memo, useEffect, useMemo, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { Link } from "react-router";
import { type FlowNodeEntity, getNodeForm } from "@flowgram.ai/fixed-layout-editor";
import { IconChevronRight, IconCircleMinus, IconPlus } from "@tabler/icons-react";
import { useControllableValue, useMount } from "ahooks";
import {
  type AnchorProps,
  AutoComplete,
  Button,
  Divider,
  Flex,
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
import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import ACMEDns01ProviderSelect from "@/components/provider/ACMEDns01ProviderSelect";
import ACMEHttp01ProviderSelect from "@/components/provider/ACMEHttp01ProviderSelect";
import CAProviderSelect from "@/components/provider/CAProviderSelect";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { ACME_DNS01_PROVIDERS, ACME_HTTP01_PROVIDERS, acmeDns01ProvidersMap, acmeHttp01ProvidersMap, caProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForBizApply, defaultNodeConfigForBizApply } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { useContactEmailsStore } from "@/stores/contact";
import { validDomainName, validIPv4Address, validIPv6Address } from "@/utils/validators";

import { FormNestedFieldsContextProvider, NodeFormContextProvider } from "./_context";
import BizApplyNodeConfigFieldsProviderAliyunESA from "./BizApplyNodeConfigFieldsProviderAliyunESA";
import BizApplyNodeConfigFieldsProviderAWSRoute53 from "./BizApplyNodeConfigFieldsProviderAWSRoute53";
import BizApplyNodeConfigFieldsProviderHuaweiCloudDNS from "./BizApplyNodeConfigFieldsProviderHuaweiCloudDNS";
import BizApplyNodeConfigFieldsProviderJDCloudDNS from "./BizApplyNodeConfigFieldsProviderJDCloudDNS";
import BizApplyNodeConfigFieldsProviderLocal from "./BizApplyNodeConfigFieldsProviderLocal";
import BizApplyNodeConfigFieldsProviderSSH from "./BizApplyNodeConfigFieldsProviderSSH";
import BizApplyNodeConfigFieldsProviderTencentCloudEO from "./BizApplyNodeConfigFieldsProviderTencentCloudEO";
import { NodeType } from "../nodes/typings";

const MULTIPLE_INPUT_SEPARATOR = ";";

const CHALLENGE_TYPE_DNS01 = "dns-01";
const CHALLENGE_TYPE_HTTP01 = "http-01";

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
    return caProvidersMap.get(fieldCAProvider)?.provider === option.provider;
  };

  const initialValues = useMemo(() => {
    return getNodeForm(node)?.getValueIn("config") as WorkflowNodeConfigForBizApply | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizApplyConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldChallengeType = Form.useWatch<string>("challengeType", { form: formInst, preserve: true });
  const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch<string>("providerAccessId", { form: formInst, preserve: true });
  const fieldCAProvider = Form.useWatch<string>("caProvider", { form: formInst, preserve: true });
  const fieldCAProviderAccessId = Form.useWatch<string>("caProviderAccessId", { form: formInst, preserve: true });

  const NestedProviderConfigFields = useMemo(() => {
    /*
      注意：如果追加新的子组件，请保持以 ASCII 排序。
      NOTICE: If you add new child component, please keep ASCII order.
      */
    switch (fieldChallengeType) {
      case CHALLENGE_TYPE_DNS01:
        {
          switch (fieldProvider) {
            case ACME_DNS01_PROVIDERS.ALIYUN_ESA: {
              return BizApplyNodeConfigFieldsProviderAliyunESA;
            }
            case ACME_DNS01_PROVIDERS.AWS:
            case ACME_DNS01_PROVIDERS.AWS_ROUTE53: {
              return BizApplyNodeConfigFieldsProviderAWSRoute53;
            }
            case ACME_DNS01_PROVIDERS.HUAWEICLOUD:
            case ACME_DNS01_PROVIDERS.HUAWEICLOUD_DNS: {
              return BizApplyNodeConfigFieldsProviderHuaweiCloudDNS;
            }
            case ACME_DNS01_PROVIDERS.JDCLOUD:
            case ACME_DNS01_PROVIDERS.JDCLOUD_DNS: {
              return BizApplyNodeConfigFieldsProviderJDCloudDNS;
            }
            case ACME_DNS01_PROVIDERS.TENCENTCLOUD_EO: {
              return BizApplyNodeConfigFieldsProviderTencentCloudEO;
            }
          }
        }
        break;

      case CHALLENGE_TYPE_HTTP01:
        switch (fieldProvider) {
          case ACME_HTTP01_PROVIDERS.DOCKERHOST: {
            return BizApplyNodeConfigFieldsProviderSSH;
          }
          case ACME_HTTP01_PROVIDERS.LOCAL: {
            return BizApplyNodeConfigFieldsProviderLocal;
          }
          case ACME_HTTP01_PROVIDERS.SSH: {
            return BizApplyNodeConfigFieldsProviderSSH;
          }
        }
        break;
    }
  }, [fieldChallengeType, fieldProvider]);

  const [showProviderAccess, setShowProviderAccess] = useState(false);
  useEffect(() => {
    // 内置的质询提供商（如本地主机）无需显示授权信息字段
    switch (fieldChallengeType) {
      case CHALLENGE_TYPE_DNS01:
        {
          if (fieldProvider) {
            const provider = acmeDns01ProvidersMap.get(fieldProvider);
            setShowProviderAccess(!provider?.builtin);
          } else {
            setShowProviderAccess(false);
          }
        }
        break;

      case CHALLENGE_TYPE_HTTP01:
        {
          if (fieldProvider) {
            const provider = acmeHttp01ProvidersMap.get(fieldProvider);
            setShowProviderAccess(!provider?.builtin);
          } else {
            setShowProviderAccess(false);
          }
        }
        break;

      default:
        {
          setShowProviderAccess(false);
        }
        break;
    }
  }, [fieldChallengeType, fieldProvider]);

  const [showCAProviderAccess, setShowCAProviderAccess] = useState(false);
  useEffect(() => {
    // 内置的 CA 提供商（如 Let's Encrypt）无需显示授权信息字段
    if (fieldCAProvider) {
      const provider = caProvidersMap.get(fieldCAProvider);
      setShowCAProviderAccess(!provider?.builtin);
    } else {
      setShowCAProviderAccess(false);
    }
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

  const handleChallengeTypeChange = (value: string) => {
    const resetFieldIfInvalid = (field: keyof z.infer<typeof formSchema>) => {
      const fieldSchame = formSchema.pick({ [field]: true });
      const fieldValue = formInst.getFieldValue(field);
      if (!fieldSchame.safeParse({ [field]: fieldValue }).success) {
        formInst.setFieldValue(field, void 0);
      }
    };

    switch (value) {
      case CHALLENGE_TYPE_DNS01:
        {
          formInst.setFieldValue("provider", void 0);
          formInst.setFieldValue("providerAccessId", void 0);
          formInst.setFieldValue("providerConfig", void 0);
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

  const handleCAProviderSelect = (value?: string | undefined) => {
    // 切换 CA 提供商时联动授权信息
    if (value == null || value === "") {
      formInst.setFieldValue("caProvider", void 0);
      formInst.setFieldValue("caProviderAccessId", void 0);
    } else if (value === initialValues?.caProvider) {
      formInst.setFieldValue("caProviderAccessId", initialValues?.caProviderAccessId);
    } else {
      if (caProvidersMap.get(fieldCAProvider)?.provider !== caProvidersMap.get(value!)?.provider) {
        formInst.setFieldValue("caProviderAccessId", void 0);
      }
    }
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item
            name="domains"
            dependencies={["challengeType"]}
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

          <Form.Item
            name="contactEmail"
            label={t("workflow_node.apply.form.contact_email.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.contact_email.tooltip") }}></span>}
          >
            <InternalEmailInput placeholder={t("workflow_node.apply.form.contact_email.placeholder")} />
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
            label={t("workflow_node.apply.form.challenge_type.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.challenge_type.tooltip") }}></span>}
          >
            <Radio.Group block onChange={(e) => handleChallengeTypeChange(e.target.value)}>
              <Radio.Button value={CHALLENGE_TYPE_DNS01}>DNS-01</Radio.Button>
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
            <div className="absolute -top-[6px] right-0 -translate-y-full">
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
            {NestedProviderConfigFields && <NestedProviderConfigFields />}
          </FormNestedFieldsContextProvider>
        </div>

        <div id="certificate" data-anchor="certificate">
          <Divider size="small">
            <Typography.Text className="text-xs font-normal" type="secondary">
              {t("workflow_node.apply.form_anchor.certificate.title")}
            </Typography.Text>
          </Divider>

          <Form.Item name="keyAlgorithm" label={t("workflow_node.apply.form.key_algorithm.label")} rules={[formRule]}>
            <Select
              options={["RSA2048", "RSA3072", "RSA4096", "RSA8192", "EC256", "EC384"].map((e) => ({
                label: e,
                value: e,
              }))}
              placeholder={t("workflow_node.apply.form.key_algorithm.placeholder")}
            />
          </Form.Item>

          <Form.Item className="relative" label={t("workflow_node.apply.form.ca_provider.label")}>
            <div className="absolute -top-[6px] right-0 -translate-y-full">
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
            <div className="absolute -top-[6px] right-0 -translate-y-full">
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
            <InternalValidityLifetimeInput placeholder={t("workflow_node.apply.form.validity_lifetime.placeholder")} />
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
              options={["classic", "tlsserver", "shortlived"].map((s) => ({ value: s }))}
              placeholder={t("workflow_node.apply.form.acme_profile.placeholder")}
              filterOption={(inputValue, option) => option!.value.toLowerCase().includes(inputValue.toLowerCase())}
            />
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
              addonAfter={t("workflow_node.apply.form.dns_propagation_wait.unit")}
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
              addonAfter={t("workflow_node.apply.form.dns_propagation_timeout.unit")}
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
              addonAfter={t("workflow_node.apply.form.dns_ttl.unit")}
            />
          </Form.Item>

          <Form.Item
            name="disableFollowCNAME"
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

          <Form.Item
            label={t("workflow_node.apply.form.skip_before_expiry_days.label")}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.skip_before_expiry_days.tooltip") }}></span>}
          >
            <Flex align="center" gap={8} wrap="wrap">
              <div>{t("workflow_node.apply.form.skip_before_expiry_days.prefix")}</div>
              <Form.Item name="skipBeforeExpiryDays" noStyle rules={[formRule]}>
                <InputNumber
                  className="w-24"
                  min={1}
                  max={365}
                  placeholder={t("workflow_node.apply.form.skip_before_expiry_days.placeholder")}
                  addonAfter={t("workflow_node.apply.form.skip_before_expiry_days.unit")}
                />
              </Form.Item>
              <div>{t("workflow_node.apply.form.skip_before_expiry_days.suffix")}</div>
            </Flex>
          </Form.Item>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const InternalEmailInput = memo(
  ({ disabled, placeholder, ...props }: { disabled?: boolean; placeholder?: string; value?: string; onChange?: (value: string) => void }) => {
    const { emails, fetchEmails, removeEmail } = useContactEmailsStore();
    useMount(() => fetchEmails(false));

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
        filterOption
        options={options}
        placeholder={placeholder}
        showSearch
        value={value}
        onChange={handleChange}
        onSearch={handleSearch}
      />
    );
  }
);

const InternalValidityLifetimeInput = memo(
  ({ disabled, placeholder, ...props }: { disabled?: boolean; placeholder?: string; value?: string; onChange?: (value: string) => void }) => {
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
          placeholder={placeholder}
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
  }
);

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
    domains: "",
    contactEmail: "",
    ...defaultNodeConfigForBizApply(),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      domains: z.string(t("workflow_node.apply.form.domains.placeholder")).refine((v) => {
        if (!v) return false;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => validDomainName(e, { allowWildcard: true }));
      }, t("common.errmsg.domain_invalid")),
      contactEmail: z.email(t("common.errmsg.email_invalid")),
      challengeType: z.string(t("workflow_node.apply.form.challenge_type.placeholder")).nonempty(t("workflow_node.apply.form.challenge_type.placeholder")),
      provider: z.string(t("workflow_node.apply.form.provider.placeholder")).nonempty(t("workflow_node.apply.form.provider.placeholder")),
      providerAccessId: z.string(t("workflow_node.apply.form.provider_access.placeholder")).nullish(),
      providerConfig: z.any().nullish(),
      caProvider: z.string().nullish(),
      caProviderAccessId: z.string().nullish(),
      caProviderConfig: z.any().nullish(),
      keyAlgorithm: z.string(t("workflow_node.apply.form.key_algorithm.placeholder")).nonempty(t("workflow_node.apply.form.key_algorithm.placeholder")),
      nameservers: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;

          return String(v)
            .split(MULTIPLE_INPUT_SEPARATOR)
            .every((e) => validIPv4Address(e) || validIPv6Address(e) || validDomainName(e));
        }, t("common.errmsg.host_invalid")),
      dnsPropagationWait: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z
          .number()
          .int(t("workflow_node.apply.form.dns_propagation_wait.placeholder"))
          .gte(0, t("workflow_node.apply.form.dns_propagation_wait.placeholder"))
          .nullish()
      ),
      dnsPropagationTimeout: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z
          .number()
          .int(t("workflow_node.apply.form.dns_propagation_timeout.placeholder"))
          .gte(1, t("workflow_node.apply.form.dns_propagation_timeout.placeholder"))
          .nullish()
      ),
      dnsTTL: z.preprocess(
        (v) => (v == null || v === "" ? void 0 : Number(v)),
        z.number().int(t("workflow_node.apply.form.dns_ttl.placeholder")).gte(1, t("workflow_node.apply.form.dns_ttl.placeholder")).nullish()
      ),
      validityLifetime: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return /^\d+[d|h]$/.test(v) && parseInt(v) > 0;
        }, t("workflow_node.apply.form.validity_lifetime.placeholder")),
      acmeProfile: z.string().nullish(),
      disableFollowCNAME: z.boolean().nullish(),
      disableARI: z.boolean().nullish(),
      skipBeforeExpiryDays: z.preprocess(
        (v) => Number(v),
        z
          .number()
          .int(t("workflow_node.apply.form.skip_before_expiry_days.placeholder"))
          .gte(1, t("workflow_node.apply.form.skip_before_expiry_days.placeholder"))
      ),
    })
    .superRefine((values, ctx) => {
      if (values.domains) {
        if (values.challengeType === CHALLENGE_TYPE_HTTP01 && values.domains.includes("*")) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.apply.form.domains.errmsg.no_wildcard_in_http01"),
            path: ["domains"],
          });
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
