import { forwardRef, memo, useEffect, useImperativeHandle, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import {
  CloseOutlined as CloseOutlinedIcon,
  PlusOutlined as PlusOutlinedIcon,
  QuestionCircleOutlined as QuestionCircleOutlinedIcon,
  RightOutlined as RightOutlinedIcon,
} from "@ant-design/icons";
import { useControllableValue } from "ahooks";
import { AutoComplete, Button, Divider, Flex, Form, type FormInstance, Input, InputNumber, Select, Switch, Tooltip, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import AccessEditModal from "@/components/access/AccessEditModal";
import AccessSelect from "@/components/access/AccessSelect";
import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import ACMEDns01ProviderSelect from "@/components/provider/ACMEDns01ProviderSelect";
import CAProviderSelect from "@/components/provider/CAProviderSelect";
import Show from "@/components/Show";
import { ACCESS_USAGES, ACME_DNS01_PROVIDERS, accessProvidersMap, acmeDns01ProvidersMap, caProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForApply, defaultNodeConfigForApply } from "@/domain/workflow";
import { useAntdForm, useAntdFormName, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { useContactEmailsStore } from "@/stores/contact";
import { validDomainName, validIPv4Address, validIPv6Address } from "@/utils/validators";

import ApplyNodeConfigFormAliyunESAConfig from "./ApplyNodeConfigFormAliyunESAConfig";
import ApplyNodeConfigFormAWSRoute53Config from "./ApplyNodeConfigFormAWSRoute53Config";
import ApplyNodeConfigFormHuaweiCloudDNSConfig from "./ApplyNodeConfigFormHuaweiCloudDNSConfig";
import ApplyNodeConfigFormJDCloudDNSConfig from "./ApplyNodeConfigFormJDCloudDNSConfig";
import ApplyNodeConfigFormTencentCloudEOConfig from "./ApplyNodeConfigFormTencentCloudEOConfig";

type ApplyNodeConfigFormFieldValues = Partial<WorkflowNodeConfigForApply>;

export type ApplyNodeConfigFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: ApplyNodeConfigFormFieldValues;
  onValuesChange?: (values: ApplyNodeConfigFormFieldValues) => void;
};

export type ApplyNodeConfigFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<ApplyNodeConfigFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<ApplyNodeConfigFormFieldValues>["resetFields"];
  validateFields: FormInstance<ApplyNodeConfigFormFieldValues>["validateFields"];
};

const MULTIPLE_INPUT_SEPARATOR = ";";

const initFormModel = (): ApplyNodeConfigFormFieldValues => {
  return defaultNodeConfigForApply();
};

const ApplyNodeConfigForm = forwardRef<ApplyNodeConfigFormInstance, ApplyNodeConfigFormProps>(
  ({ className, style, disabled, initialValues, onValuesChange }, ref) => {
    const { t } = useTranslation();

    const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));

    const formSchema = z.object({
      domains: z.string(t("workflow_node.apply.form.domains.placeholder")).refine((v) => {
        if (!v) return false;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => validDomainName(e, { allowWildcard: true }));
      }, t("common.errmsg.domain_invalid")),
      contactEmail: z.email(t("common.errmsg.email_invalid")),
      challengeType: z.string().nullish(),
      provider: z.string(t("workflow_node.apply.form.provider.placeholder")).nonempty(t("workflow_node.apply.form.provider.placeholder")),
      providerAccessId: z.string(t("workflow_node.apply.form.provider_access.placeholder")).min(1, t("workflow_node.apply.form.provider_access.placeholder")),
      providerConfig: z.any().nullish(),
      caProvider: z.string(t("workflow_node.apply.form.ca_provider.placeholder")).nullish(),
      caProviderAccessId: z
        .string(t("workflow_node.apply.form.ca_provider_access.placeholder"))
        .nullish()
        .refine((v) => {
          if (!fieldCAProvider) return true;

          const provider = caProvidersMap.get(fieldCAProvider);
          return !!provider?.builtin || !!v;
        }, t("workflow_node.apply.form.ca_provider_access.placeholder")),
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
        (v) => (v == null || v === "" ? undefined : Number(v)),
        z
          .number()
          .int(t("workflow_node.apply.form.dns_propagation_wait.placeholder"))
          .gte(0, t("workflow_node.apply.form.dns_propagation_wait.placeholder"))
          .nullish()
      ),
      dnsPropagationTimeout: z.preprocess(
        (v) => (v == null || v === "" ? undefined : Number(v)),
        z
          .number()
          .int(t("workflow_node.apply.form.dns_propagation_timeout.placeholder"))
          .gte(1, t("workflow_node.apply.form.dns_propagation_timeout.placeholder"))
          .nullish()
      ),
      dnsTTL: z.preprocess(
        (v) => (v == null || v === "" ? undefined : Number(v)),
        z.number().int(t("workflow_node.apply.form.dns_ttl.placeholder")).gte(1, t("workflow_node.apply.form.dns_ttl.placeholder")).nullish()
      ),
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
    });
    const formRule = createSchemaFieldRule(formSchema);
    const { form: formInst, formProps } = useAntdForm({
      name: "workflowNodeApplyConfigForm",
      initialValues: initialValues ?? initFormModel(),
    });

    const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });
    const fieldProviderAccessId = Form.useWatch<string>("providerAccessId", formInst);
    const fieldCAProvider = Form.useWatch<string>("caProvider", formInst);

    const [showProvider, setShowProvider] = useState(false);
    useEffect(() => {
      // 通常情况下每个授权信息只对应一个 DNS 提供商，此时无需显示 DNS 提供商字段；
      // 如果对应多个（如 AWS 的 Route53、Lightsail，阿里云的 DNS、ESA，腾讯云的 DNS、EdgeOne 等），则显示。
      if (fieldProviderAccessId) {
        const access = accesses.find((e) => e.id === fieldProviderAccessId);
        const providers = Array.from(acmeDns01ProvidersMap.values()).filter((e) => e.provider === access?.provider);
        setShowProvider(providers.length > 1);
      } else {
        setShowProvider(false);
      }
    }, [accesses, fieldProviderAccessId]);

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

    const [nestedFormInst] = Form.useForm();
    const nestedFormName = useAntdFormName({ form: nestedFormInst, name: "workflowNodeApplyConfigFormProviderConfigForm" });
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
        case ACME_DNS01_PROVIDERS.ALIYUN_ESA:
          return <ApplyNodeConfigFormAliyunESAConfig {...nestedFormProps} />;
        case ACME_DNS01_PROVIDERS.AWS:
        case ACME_DNS01_PROVIDERS.AWS_ROUTE53:
          return <ApplyNodeConfigFormAWSRoute53Config {...nestedFormProps} />;
        case ACME_DNS01_PROVIDERS.HUAWEICLOUD:
        case ACME_DNS01_PROVIDERS.HUAWEICLOUD_DNS:
          return <ApplyNodeConfigFormHuaweiCloudDNSConfig {...nestedFormProps} />;
        case ACME_DNS01_PROVIDERS.JDCLOUD:
        case ACME_DNS01_PROVIDERS.JDCLOUD_DNS:
          return <ApplyNodeConfigFormJDCloudDNSConfig {...nestedFormProps} />;
        case ACME_DNS01_PROVIDERS.TENCENTCLOUD_EO:
          return <ApplyNodeConfigFormTencentCloudEOConfig {...nestedFormProps} />;
      }
    }, [disabled, initialValues?.providerConfig, fieldProvider, nestedFormInst, nestedFormName]);

    const handleProviderSelect = (value: string) => {
      if (fieldProvider === value) return;

      // 切换 DNS 提供商时联动授权信息
      if (initialValues?.provider === value) {
        formInst.setFieldValue("providerAccessId", initialValues?.providerAccessId);
        onValuesChange?.(formInst.getFieldsValue(true));
      } else {
        if (acmeDns01ProvidersMap.get(fieldProvider)?.provider !== acmeDns01ProvidersMap.get(value)?.provider) {
          formInst.setFieldValue("providerAccessId", undefined);
          onValuesChange?.(formInst.getFieldsValue(true));
        }
      }
    };

    const handleProviderAccessSelect = (value: string) => {
      // 切换授权信息时联动 DNS 提供商
      const access = accesses.find((access) => access.id === value);
      const provider = Array.from(acmeDns01ProvidersMap.values()).find((provider) => provider.provider === access?.provider);
      if (fieldProvider !== provider?.type) {
        formInst.setFieldValue("provider", provider?.type);
        onValuesChange?.(formInst.getFieldsValue(true));
      }
    };

    const handleCAProviderSelect = (value?: string | undefined) => {
      // 切换 CA 提供商时联动授权信息
      if (value === "") {
        setTimeout(() => {
          formInst.setFieldValue("caProvider", undefined);
          formInst.setFieldValue("caProviderAccessId", undefined);
          onValuesChange?.(formInst.getFieldsValue(true));
        }, 1);
      } else if (initialValues?.caProvider === value) {
        formInst.setFieldValue("caProviderAccessId", initialValues?.caProviderAccessId);
        onValuesChange?.(formInst.getFieldsValue(true));
      } else {
        if (caProvidersMap.get(fieldCAProvider)?.provider !== caProvidersMap.get(value!)?.provider) {
          formInst.setFieldValue("caProviderAccessId", undefined);
          onValuesChange?.(formInst.getFieldsValue(true));
        }
      }
    };

    const handleFormProviderChange = (name: string) => {
      if (name === nestedFormName) {
        formInst.setFieldValue("providerConfig", nestedFormInst.getFieldsValue());
        onValuesChange?.(formInst.getFieldsValue(true));
      }
    };

    const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
      onValuesChange?.(values as ApplyNodeConfigFormFieldValues);
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
      } as ApplyNodeConfigFormInstance;
    });

    return (
      <Form.Provider onFormChange={handleFormProviderChange}>
        <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
          <Form.Item
            name="domains"
            label={t("workflow_node.apply.form.domains.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.domains.tooltip") }}></span>}
          >
            <MultipleSplitValueInput
              modalTitle={t("workflow_node.apply.form.domains.multiple_input_modal.title")}
              placeholder={t("workflow_node.apply.form.domains.placeholder")}
              placeholderInModal={t("workflow_node.apply.form.domains.multiple_input_modal.placeholder")}
              splitOptions={{ trim: true, removeEmpty: true }}
            />
          </Form.Item>

          <Form.Item
            name="contactEmail"
            label={t("workflow_node.apply.form.contact_email.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.contact_email.tooltip") }}></span>}
          >
            <EmailInput placeholder={t("workflow_node.apply.form.contact_email.placeholder")} />
          </Form.Item>

          <Form.Item name="challengeType" label={t("workflow_node.apply.form.challenge_type.label")} rules={[formRule]} hidden>
            <Select
              options={["DNS-01"].map((e) => ({
                label: e,
                value: e.toLowerCase(),
              }))}
              placeholder={t("workflow_node.apply.form.challenge_type.placeholder")}
            />
          </Form.Item>

          <Form.Item name="provider" label={t("workflow_node.apply.form.provider.label")} hidden={!showProvider} rules={[formRule]}>
            <ACMEDns01ProviderSelect
              disabled={!showProvider}
              filter={(record) => {
                if (fieldProviderAccessId) {
                  return accesses.find((e) => e.id === fieldProviderAccessId)?.provider === record.provider;
                }

                return true;
              }}
              placeholder={t("workflow_node.apply.form.provider.placeholder")}
              showSearch
              onSelect={handleProviderSelect}
            />
          </Form.Item>

          <Form.Item noStyle>
            <label className="mb-1 block">
              <div className="flex w-full items-center justify-between gap-4">
                <div className="max-w-full grow truncate">
                  <span>{t("workflow_node.apply.form.provider_access.label")}</span>
                  <Tooltip title={t("workflow_node.apply.form.provider_access.tooltip")}>
                    <Typography.Text className="ms-1" type="secondary">
                      <QuestionCircleOutlinedIcon />
                    </Typography.Text>
                  </Tooltip>
                </div>
                <div className="text-right">
                  <AccessEditModal
                    scene="add"
                    trigger={
                      <Button size="small" type="link">
                        {t("workflow_node.apply.form.provider_access.button")}
                        <PlusOutlinedIcon className="text-xs" />
                      </Button>
                    }
                    usage="both-dns-hosting"
                    afterSubmit={(record) => {
                      const provider = accessProvidersMap.get(record.provider);
                      if (provider?.usages?.includes(ACCESS_USAGES.DNS)) {
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
                  if (record.reserve) return false;

                  const provider = accessProvidersMap.get(record.provider);
                  return !!provider?.usages?.includes(ACCESS_USAGES.DNS);
                }}
                placeholder={t("workflow_node.apply.form.provider_access.placeholder")}
                showSearch
                onChange={handleProviderAccessSelect}
              />
            </Form.Item>
          </Form.Item>
        </Form>

        {nestedFormEl}

        <Divider size="small">
          <Typography.Text className="text-xs font-normal" type="secondary">
            {t("workflow_node.apply.form.certificate_config.label")}
          </Typography.Text>
        </Divider>

        <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
          <Form.Item noStyle>
            <label className="mb-1 block">
              <div className="flex w-full items-center justify-between gap-4">
                <div className="max-w-full grow truncate">
                  <span>{t("workflow_node.apply.form.ca_provider.label")}</span>
                </div>
                <div className="text-right">
                  <Show when={!fieldCAProvider}>
                    <Link className="ant-typography" to="/settings/ssl-provider" target="_blank">
                      <Button size="small" type="link">
                        {t("workflow_node.apply.form.ca_provider.button")}
                        <RightOutlinedIcon className="text-xs" />
                      </Button>
                    </Link>
                  </Show>
                </div>
              </div>
            </label>
            <Form.Item name="caProvider" rules={[formRule]}>
              <CAProviderSelect
                allowClear
                placeholder={t("workflow_node.apply.form.ca_provider.placeholder")}
                showSearch
                onSelect={handleCAProviderSelect}
                onClear={handleCAProviderSelect}
              />
            </Form.Item>
          </Form.Item>

          <Form.Item hidden={!showCAProviderAccess} noStyle>
            <label className="mb-1 block">
              <div className="flex w-full items-center justify-between gap-4">
                <div className="max-w-full grow truncate">
                  <span>{t("workflow_node.apply.form.ca_provider_access.label")}</span>
                </div>
                <div className="text-right">
                  <AccessEditModal
                    data={{ provider: caProvidersMap.get(fieldCAProvider!)?.provider }}
                    scene="add"
                    trigger={
                      <Button size="small" type="link">
                        {t("workflow_node.apply.form.ca_provider_access.button")}
                        <PlusOutlinedIcon className="text-xs" />
                      </Button>
                    }
                    usage="ca-only"
                    afterSubmit={(record) => {
                      const provider = accessProvidersMap.get(record.provider);
                      if (provider?.usages?.includes(ACCESS_USAGES.CA)) {
                        formInst.setFieldValue("caProviderAccessId", record.id);
                      }
                    }}
                  />
                </div>
              </div>
            </label>
            <Form.Item name="caProviderAccessId" rules={[formRule]}>
              <AccessSelect
                filter={(record) => {
                  if (record.reserve !== "ca") return false;
                  if (fieldCAProvider) return caProvidersMap.get(fieldCAProvider)?.provider === record.provider;

                  const provider = accessProvidersMap.get(record.provider);
                  return !!provider?.usages?.includes(ACCESS_USAGES.CA);
                }}
                placeholder={t("workflow_node.apply.form.ca_provider_access.placeholder")}
                showSearch
              />
            </Form.Item>
          </Form.Item>

          <Form.Item name="keyAlgorithm" label={t("workflow_node.apply.form.key_algorithm.label")} rules={[formRule]}>
            <Select
              options={["RSA2048", "RSA3072", "RSA4096", "RSA8192", "EC256", "EC384"].map((e) => ({
                label: e,
                value: e,
              }))}
              placeholder={t("workflow_node.apply.form.key_algorithm.placeholder")}
            />
          </Form.Item>

          <Form.Item
            name="acmeProfile"
            label={t("workflow_node.apply.form.acme_profile.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.acme_profile.tooltip") }}></span>}
          >
            <AutoComplete
              allowClear
              options={["classic", "tlsserver", "shortlived"].map((value) => ({ value }))}
              placeholder={t("workflow_node.apply.form.acme_profile.placeholder")}
              filterOption={(inputValue, option) => option!.value.toLowerCase().includes(inputValue.toLowerCase())}
            />
          </Form.Item>
        </Form>

        <Divider size="small">
          <Typography.Text className="text-xs font-normal" type="secondary">
            {t("workflow_node.apply.form.advanced_config.label")}
          </Typography.Text>
        </Divider>

        <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
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
              splitOptions={{ trim: true, removeEmpty: true }}
            />
          </Form.Item>

          <Form.Item
            name="dnsPropagationWait"
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
            label={t("workflow_node.apply.form.dns_ttl.label")}
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
        </Form>

        <Divider size="small">
          <Typography.Text className="text-xs font-normal" type="secondary">
            {t("workflow_node.apply.form.strategy_config.label")}
          </Typography.Text>
        </Divider>

        <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
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
        </Form>
      </Form.Provider>
    );
  }
);

const EmailInput = memo(
  ({ disabled, placeholder, ...props }: { disabled?: boolean; placeholder?: string; value?: string; onChange?: (value: string) => void }) => {
    const { emails, fetchEmails, removeEmail } = useContactEmailsStore();
    useEffect(() => {
      fetchEmails();
    }, []);

    const [value, setValue] = useControllableValue<string>(props, {
      valuePropName: "value",
      defaultValuePropName: "defaultValue",
      trigger: "onChange",
    });

    const [inputValue, setInputValue] = useState<string>();

    const renderOptionLabel = (email: string, removable: boolean = false) => (
      <div className="flex items-center gap-2 overflow-hidden">
        <span className="flex-1 overflow-hidden truncate">{email}</span>
        {removable && (
          <Button
            color="default"
            disabled={disabled}
            icon={<CloseOutlinedIcon />}
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

export default memo(ApplyNodeConfigForm);
