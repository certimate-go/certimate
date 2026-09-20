import { useEffect, useMemo } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { IconPlus } from "@tabler/icons-react";
import { type AnchorProps, Button, Form, type FormInstance, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AccessEditDrawer from "@/components/access/AccessEditDrawer";
import AccessSelect from "@/components/access/AccessSelect";
import FileTextInput from "@/components/FileTextInput";
import CertsourceProviderSelect from "@/components/provider/CertsourceProviderSelect";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { type AccessModel } from "@/domain/access";
import { certsourceProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForBizUpload, defaultNodeConfigForBizUpload } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { getCertificateSubjectAltNames as getX509SubjectAltNames, validatePEMCertificate, validatePEMPrivateKey } from "@/utils/x509";

import { FormNestedFieldsContextProvider, NodeFormContextProvider } from "./_context";
import BizUploadNodeConfigFieldsProvider from "./BizUploadNodeConfigFieldsProvider";
import { NodeType } from "../nodes/typings";

export interface BizUploadNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const UPLOAD_SOURCE_FORM = "form" as const;
const UPLOAD_SOURCE_LOCAL = "local" as const;
const UPLOAD_SOURCE_URL = "url" as const;
const UPLOAD_SOURCE_PROVIDER = "provider" as const;

const BizUploadNodeConfigForm = ({ node, ...props }: BizUploadNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizUpload) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizUpload}`);
  }

  const { i18n, t } = useTranslation();

  const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));
  const accessOptionFilter = (_: string, option: AccessModel) => {
    if (option.reserve) return false;
    return certsourceProvidersMap.get(fieldProvider ?? "")?.provider === option.provider;
  };

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBizUpload | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizUploadConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldSource = Form.useWatch("source", { form: formInst, preserve: true });
  const fieldProvider = Form.useWatch("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch("providerAccessId", { form: formInst, preserve: true });
  const fieldCertificate = Form.useWatch("certificate", { form: formInst, preserve: true });
  const fieldName = useMemo(() => {
    if (!fieldSource || fieldSource === UPLOAD_SOURCE_FORM) {
      return fieldCertificate ? getX509SubjectAltNames(fieldCertificate).join(";") : void 0;
    }
    return void 0;
  }, [fieldSource, fieldCertificate]);

  const renderNestedFieldProviderComponent = BizUploadNodeConfigFieldsProvider.useComponent(fieldProvider ?? "", {});

  useEffect(() => {
    if (fieldSource !== UPLOAD_SOURCE_PROVIDER) return;

    // 如果未选择提供商，则清空授权信息
    if (!fieldProvider && fieldProviderAccessId) {
      formInst.setFieldValue("providerAccessId", void 0);
      return;
    }

    // 如果已选择提供商只有一个授权信息，则自动选择该授权信息
    if (fieldProvider && !fieldProviderAccessId) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilter(access.provider, access))
        .filter((access) => certsourceProvidersMap.get(fieldProvider)?.provider === access.provider);
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("providerAccessId", availableAccesses[0].id);
      }
    }
  }, [fieldSource, fieldProvider, fieldProviderAccessId]);

  const handleSourceChange = (value: string) => {
    if (value === initialValues?.source) {
      formInst.resetFields(["certificate", "privateKey"]);
    } else {
      setTimeout(() => {
        formInst.setFieldValue("certificate", "");
        formInst.setFieldValue("privateKey", "");
      }, 0);
    }
  };

  const handleProviderSelect = (value?: string | undefined) => {
    // 切换提供商时重置表单，避免其他提供商的配置字段残留
    formInst.setFieldValue("providerAccessId", void 0);
    if (initialValues?.provider === value) {
      formInst.resetFields(["providerConfig"]);
    } else {
      formInst.setFieldValue("providerConfig", void 0);
    }
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item name="source" label={t("workflow_node.upload.form.source.label")} rules={[formRule]}>
            <Radio.Group block onChange={(e) => handleSourceChange(e.target.value)}>
              <Radio.Button value={UPLOAD_SOURCE_FORM}>{t("workflow_node.upload.form.source.option.form.label")}</Radio.Button>
              <Radio.Button value={UPLOAD_SOURCE_LOCAL}>{t("workflow_node.upload.form.source.option.local.label")}</Radio.Button>
              <Radio.Button value={UPLOAD_SOURCE_URL}>{t("workflow_node.upload.form.source.option.url.label")}</Radio.Button>
              <Radio.Button value={UPLOAD_SOURCE_PROVIDER}>{t("workflow_node.upload.form.source.option.provider.label")}</Radio.Button>
            </Radio.Group>
          </Form.Item>

          <Show when={fieldSource === UPLOAD_SOURCE_FORM}>
            <Form.Item label={t("workflow_node.upload.form.name.label")}>
              <Input placeholder={t("workflow_node.upload.form.name.placeholder")} readOnly value={fieldName} variant="filled" />
            </Form.Item>

            <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate_pem.label")} rules={[formRule]}>
              <FileTextInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("workflow_node.upload.form.certificate_pem.placeholder")} />
            </Form.Item>

            <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key_pem.label")} rules={[formRule]}>
              <FileTextInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("workflow_node.upload.form.private_key_pem.placeholder")} />
            </Form.Item>
          </Show>

          <Show when={fieldSource === UPLOAD_SOURCE_LOCAL}>
            <Form.Item>
              <Tips message={t("workflow_node.upload.form.guide")} />
            </Form.Item>

            <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate_path.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.certificate_path.placeholder")} />
            </Form.Item>

            <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key_path.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.private_key_path.placeholder")} />
            </Form.Item>
          </Show>

          <Show when={fieldSource === UPLOAD_SOURCE_URL}>
            <Form.Item>
              <Tips message={t("workflow_node.upload.form.guide")} />
            </Form.Item>

            <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate_url.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.certificate_url.placeholder")} />
            </Form.Item>

            <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key_url.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.private_key_url.placeholder")} />
            </Form.Item>
          </Show>

          <Show when={fieldSource === UPLOAD_SOURCE_PROVIDER}>
            <Form.Item>
              <Tips message={t("workflow_node.upload.form.provider_guide")} />
            </Form.Item>

            <Form.Item name="provider" label={t("workflow_node.upload.form.certsource_provider.label")} rules={[formRule]}>
              <CertsourceProviderSelect
                allowClear
                placeholder={t("workflow_node.upload.form.certsource_provider.placeholder")}
                showAvailability
                showSearch
                onSelect={handleProviderSelect}
                onClear={handleProviderSelect}
              />
            </Form.Item>

            <Form.Item className="relative" label={t("workflow_node.upload.form.certsource_provider_access.label")}>
              <div className="absolute -top-1.5 right-0 -translate-y-full">
                <AccessEditDrawer
                  data={{ provider: certsourceProvidersMap.get(fieldProvider!)?.provider }}
                  mode="create"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.upload.form.certsource_provider_access.button")}
                      <IconPlus size="1.25em" />
                    </Button>
                  }
                  usage="certsource"
                  afterSubmit={(record) => {
                    if (!accessOptionFilter(record.provider, record)) return;
                    if (certsourceProvidersMap.get(fieldProvider!)?.provider !== record.provider) return;
                    formInst.setFieldValue("providerAccessId", record.id);
                  }}
                />
              </div>
              <Form.Item name="providerAccessId" dependencies={["provider"]} rules={[formRule]} noStyle>
                <AccessSelect
                  disabled={!fieldProvider}
                  placeholder={t("workflow_node.upload.form.certsource_provider_access.placeholder")}
                  showSearch
                  onFilter={accessOptionFilter}
                />
              </Form.Item>
            </Form.Item>

            <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
              {renderNestedFieldProviderComponent && <>{renderNestedFieldProviderComponent}</>}
            </FormNestedFieldsContextProvider>
          </Show>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters"].map((key) => ({
    key: key,
    title: t(`workflow_node.upload.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    source: UPLOAD_SOURCE_FORM,
    certificate: "",
    privateKey: "",
    ...(defaultNodeConfigForBizUpload() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      source: z.enum([UPLOAD_SOURCE_FORM, UPLOAD_SOURCE_LOCAL, UPLOAD_SOURCE_URL, UPLOAD_SOURCE_PROVIDER]).default(UPLOAD_SOURCE_FORM),
      name: z.string().nullish(),
      certificate: z.string().nullish(),
      privateKey: z.string().nullish(),
      provider: z.string().nullish(),
      providerAccessId: z.string().nullish(),
      providerConfig: z.any().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.source) {
        case UPLOAD_SOURCE_FORM:
          {
            const scCertificate = z.string().refine((v) => validatePEMCertificate(v), t("workflow_node.upload.form.certificate_pem.errmsg.invalid"));
            const spCertificate = scCertificate.safeParse(values.certificate);
            if (!spCertificate.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spCertificate.error).errors.join(),
                path: ["certificate"],
              });
            }

            const scPrivateKey = z.string().refine((v) => validatePEMPrivateKey(v), t("workflow_node.upload.form.private_key_pem.errmsg.invalid"));
            const spPrivateKey = scPrivateKey.safeParse(values.privateKey);
            if (!spPrivateKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPrivateKey.error).errors.join(),
                path: ["privateKey"],
              });
            }
          }
          break;

        case UPLOAD_SOURCE_LOCAL:
          {
            const scCertificate = z.string().nonempty();
            const spCertificate = scCertificate.safeParse(values.certificate);
            if (!spCertificate.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spCertificate.error).errors.join(),
                path: ["certificate"],
              });
            }

            const scPrivateKey = z.string().nonempty();
            const spPrivateKey = scPrivateKey.safeParse(values.privateKey);
            if (!spPrivateKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPrivateKey.error).errors.join(),
                path: ["privateKey"],
              });
            }
          }
          break;

        case UPLOAD_SOURCE_URL:
          {
            const scCertificate = z.url({ protocol: z.core.regexes.httpProtocol });
            const spCertificate = scCertificate.safeParse(values.certificate);
            if (!spCertificate.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spCertificate.error).errors.join(),
                path: ["certificate"],
              });
            }

            const scPrivateKey = z.url({ protocol: z.core.regexes.httpProtocol });
            const spPrivateKey = scPrivateKey.safeParse(values.privateKey);
            if (!spPrivateKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPrivateKey.error).errors.join(),
                path: ["privateKey"],
              });
            }
          }
          break;

        case UPLOAD_SOURCE_PROVIDER:
          {
            const scProvider = z.string().nonempty(t("workflow_node.upload.form.certsource_provider.placeholder"));
            const spProvider = scProvider.safeParse(values.provider);
            if (!spProvider.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spProvider.error).errors.join(),
                path: ["provider"],
              });
            }

            const scProviderAccessId = z.string().nonempty(t("workflow_node.upload.form.certsource_provider_access.placeholder"));
            const spProviderAccessId = scProviderAccessId.safeParse(values.providerAccessId);
            if (!spProviderAccessId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spProviderAccessId.error).errors.join(),
                path: ["providerAccessId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizUploadNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
