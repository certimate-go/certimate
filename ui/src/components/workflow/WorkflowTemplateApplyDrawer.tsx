import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconArrowRight } from "@tabler/icons-react";
import { useControllableValue } from "ahooks";
import { Alert, App, Button, Divider, Drawer, Form, type FormInstance, Input, Spin, Tag, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { ClientResponseError } from "pocketbase";
import { z } from "zod";

import { type ProviderSchemaEnvelope, getProviderSchema } from "@/api/providerschema";
import { instantiateWorkflowTemplate } from "@/api/workflowtemplates";
import AccessSelect from "@/components/access/AccessSelect";
import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { deploymentProvidersMap } from "@/domain/provider";
import {
  type WorkflowTemplateModel,
  type WorkflowTemplateParam,
  type WorkflowTemplateParamError,
  type WorkflowTemplateParamValue,
  findWorkflowGraphNodeById,
  sortWorkflowTemplateParams,
} from "@/domain/workflowTemplate";
import { useAntdForm } from "@/hooks";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { unwrapErrMsg } from "@/utils/error";
import { isDomain } from "@/utils/validator";

import { FormNestedFieldsContextProvider } from "./designer/forms/_context";
import BizDeployNodeConfigFieldsProvider from "./designer/forms/BizDeployNodeConfigFieldsProvider";
import SchemaConfigFields from "./designer/forms/SchemaConfigFields";

const MULTIPLE_INPUT_SEPARATOR = ";";

const deployConfigFieldName = (paramId: string) => `__deployConfig__${paramId}`;
const deployAccessFieldName = (paramId: string) => `__deployAccess__${paramId}`;

const localizeParamError = (err: WorkflowTemplateParamError, t: (key: string, values?: Record<string, unknown>) => string) => {
  switch (err.code) {
    case "required":
      return t("workflow.template.apply.errmsg.required");
    case "invalidDomain":
      return t("workflow.template.apply.errmsg.invalid_domain");
    case "accessRequired":
      return t("workflow.template.apply.errmsg.access_required");
    case "accessNotFound":
      return t("workflow.template.apply.errmsg.access_not_found");
    case "accessProviderMismatch":
      return t("workflow.template.apply.errmsg.access_provider_mismatch");
    case "providerUnavailable":
      return t("workflow.template.apply.errmsg.provider_unavailable");
    case "configInvalid":
      return t("workflow.template.apply.errmsg.config_invalid", { message: err.message });
    default:
      return err.message;
  }
};

const isAccessRelatedParamError = (err: WorkflowTemplateParamError) => {
  return ["accessRequired", "accessNotFound", "accessProviderMismatch"].includes(err.code ?? "");
};

export interface WorkflowTemplateApplyDrawerProps {
  afterClose?: () => void;
  open?: boolean;
  template?: WorkflowTemplateModel;
  onInstantiated?: (workflowId: string) => void;
  onOpenChange?: (open: boolean) => void;
}

const WorkflowTemplateApplyDrawer = ({ afterClose, template, onInstantiated, ...props }: WorkflowTemplateApplyDrawerProps) => {
  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const params = useMemo(() => sortWorkflowTemplateParams(template?.params ?? []), [template]);

  const [paramErrs, setParamErrs] = useState<WorkflowTemplateParamError[]>([]);
  const [revisionConflict, setRevisionConflict] = useState(false);
  const [submitPending, setSubmitPending] = useState(false);

  const initialValues = useMemo(() => {
    const values: Record<string, unknown> = {
      name: "",
      description: template?.description ?? "",
    };
    for (const param of params) {
      switch (param.kind) {
        case "domains":
          values[param.id] = "";
          break;
        case "deployment":
          {
            const node = findWorkflowGraphNodeById(template?.graph, param.deployNodeId);
            const nodeConfig = (node?.data?.config ?? {}) as Record<string, unknown>;
            values[deployConfigFieldName(param.id)] = (nodeConfig.providerConfig as Record<string, unknown>) ?? {};
            values[deployAccessFieldName(param.id)] = void 0;
          }
          break;
      }
    }
    return values;
  }, [template, params]);

  const formSchema = useMemo(() => {
    const shape: Record<string, z.ZodTypeAny> = {
      name: z.string().nonempty(t("workflow.template.apply.form.name.errmsg.invalid")),
      description: z.string().nullish(),
    };
    for (const param of params) {
      switch (param.kind) {
        case "domains":
          shape[param.id] = z
            .string()
            .nullish()
            .refine((v) => {
              if (!v) return false;
              return v
                .split(MULTIPLE_INPUT_SEPARATOR)
                .filter((s) => s.trim() !== "")
                .every((s) => isDomain(s.trim(), { allowWildcard: true }));
            }, t("workflow.template.apply.form.domains.errmsg.invalid"));
          break;

        case "deployment":
          {
            const node = findWorkflowGraphNodeById(template?.graph, param.deployNodeId);
            const provider = ((node?.data?.config ?? {}) as Record<string, unknown>).provider as string | undefined;
            if (provider && !deploymentProvidersMap.get(provider)?.builtin) {
              shape[deployAccessFieldName(param.id)] = z
                .string()
                .nullish()
                .refine((v) => !!v, t("workflow.template.apply.form.provider_access.errmsg.invalid"));
            } else {
              shape[deployAccessFieldName(param.id)] = z.string().nullish();
            }
            shape[deployConfigFieldName(param.id)] = z.any().nullish();
          }
          break;
      }
    }
    return z.object(shape);
  }, [params, template, t]);
  const formRule = createSchemaFieldRule(formSchema);

  const { form: formInst, formProps } = useAntdForm({
    name: "workflowTemplateApplyForm",
    initialValues,
    onSubmit: handleSubmit,
  });

  const accessOptionFilter = (provider: string) => {
    return (_value: string, option: AccessModel) => {
      if (option.reserve) return false;
      return deploymentProvidersMap.get(provider)?.provider === option.provider;
    };
  };

  const handleCancelClick = () => {
    setOpen(false);
    setParamErrs([]);
    setRevisionConflict(false);
  };

  async function handleSubmit(values: Record<string, unknown>) {
    if (!template) return;

    const reqParams: Record<string, WorkflowTemplateParamValue> = {};
    for (const param of params) {
      switch (param.kind) {
        case "domains":
          reqParams[param.id] = {
            domains: String(values[param.id] ?? "")
              .split(MULTIPLE_INPUT_SEPARATOR)
              .map((s) => s.trim())
              .filter((s) => s !== ""),
          };
          break;

        case "deployment":
          reqParams[param.id] = {
            providerAccessId: String(values[deployAccessFieldName(param.id)] ?? ""),
            providerConfig: (values[deployConfigFieldName(param.id)] as Record<string, unknown>) ?? {},
          };
          break;
      }
    }

    setSubmitPending(true);
    setParamErrs([]);
    setRevisionConflict(false);
    try {
      const resp = await instantiateWorkflowTemplate(template.id, {
        revision: template.revision,
        name: String(values.name ?? "").trim(),
        description: String(values.description ?? "").trim(),
        params: reqParams,
      });

      message.success(t("workflow.template.apply.text.created"));
      setOpen(false);
      onInstantiated?.(resp.workflowId);
    } catch (err) {
      if (err instanceof ClientResponseError && err.status === 409) {
        setRevisionConflict(true);
        return;
      }

      const errs = err instanceof ClientResponseError ? (err.response?.data?.errors as WorkflowTemplateParamError[] | undefined) : undefined;
      if (errs?.length) {
        setParamErrs(errs);

        const fields: Parameters<FormInstance["setFields"]>[0] = [];
        for (const param of params) {
          const paramErr = errs.find((e) => e.paramId === param.id);
          if (!paramErr) continue;

          switch (param.kind) {
            case "domains":
              fields.push({ name: param.id, errors: [localizeParamError(paramErr, t)] });
              break;
            case "deployment":
              if (isAccessRelatedParamError(paramErr)) {
                fields.push({ name: deployAccessFieldName(param.id), errors: [localizeParamError(paramErr, t)] });
              }
              break;
          }
        }
        if (fields.length > 0) formInst.setFields(fields);
        return;
      }

      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    } finally {
      setSubmitPending(false);
    }
  }

  return (
    <Drawer
      afterOpenChange={(visible) => {
        if (!visible) afterClose?.();
      }}
      closable
      destroyOnHidden
      footer={
        <div className="flex items-center justify-between gap-4">
          <Typography.Text type="secondary" className="text-xs">
            {t("workflow.template.apply.text.footer_notice")}
          </Typography.Text>
          <div className="flex gap-2">
            <Button onClick={handleCancelClick}>{t("common.button.cancel")}</Button>
            <Button
              disabled={submitPending}
              type="primary"
              onClick={() => {
                formInst.submit();
              }}
            >
              {t("workflow.template.apply.button.create")}
              <IconArrowRight size="1.25em" />
            </Button>
          </div>
        </div>
      }
      loading={!template}
      open={open}
      title={template ? t("workflow.template.apply.title", { name: template.name }) : t("workflow.template.apply.title_loading")}
      width="640px"
      onClose={handleCancelClick}
    >
      <Show
        when={!!template}
        fallback={
          <div className="flex h-32 items-center justify-center">
            <Spin />
          </div>
        }
      >
        <div className="pt-2">
          {revisionConflict && (
            <Alert
              className="mb-4"
              showIcon
              title={t("workflow.template.apply.revision_conflict.title")}
              description={t("workflow.template.apply.revision_conflict.description")}
              type="warning"
            />
          )}

          <Form {...formProps} clearOnDestroy={false} form={formInst} layout="vertical" preserve={true}>
            <Form.Item name="name" label={t("workflow.template.apply.form.name.label")} rules={[formRule]}>
              <Input />
            </Form.Item>
            <Form.Item name="description" label={t("workflow.template.apply.form.description.label")}>
              <Input />
            </Form.Item>

            {params.map((param) => (
              <TemplateParamGroup
                key={param.id}
                formRule={formRule}
                param={param}
                paramErr={paramErrs.find((e) => e.paramId === param.id)}
                template={template!}
                getAccessOptionFilter={accessOptionFilter}
              />
            ))}

            {params.length === 0 && <Alert className="mt-2" showIcon title={t("workflow.template.apply.text.no_params")} type="info" />}
          </Form>
        </div>
      </Show>
    </Drawer>
  );
};

type TemplateParamGroupProps = {
  formRule: ReturnType<typeof createSchemaFieldRule>;
  getAccessOptionFilter: (provider: string) => (value: string, option: AccessModel) => boolean;
  param: WorkflowTemplateParam;
  paramErr?: WorkflowTemplateParamError;
  template: WorkflowTemplateModel;
};

const TemplateParamGroup = ({ formRule, getAccessOptionFilter, param, paramErr, template }: TemplateParamGroupProps) => {
  const { t } = useTranslation();

  const node = findWorkflowGraphNodeById(template.graph, param.kind === "deployment" ? param.deployNodeId : void 0);
  const provider = ((node?.data?.config ?? {}) as Record<string, unknown>).provider as string | undefined;
  const providerInfo = provider ? deploymentProvidersMap.get(provider) : undefined;
  const providerLabel = providerInfo ? t(providerInfo.name) : provider;

  const showProviderAccess = !!provider && !providerInfo?.builtin;

  return (
    <div className="mt-4 rounded-lg border border-gray-200 p-3 dark:border-gray-700">
      <Divider className="mt-0" size="small" titlePlacement="left">
        <span className="font-medium">{param.title}</span>
        <Tag className="ms-2" color="default">
          {param.kind === "domains"
            ? t("workflow.template.apply.tag.domains")
            : `${t("workflow.template.apply.tag.deployment")}${providerLabel ? ` · ${providerLabel}` : ""}`}
        </Tag>
      </Divider>

      <Show
        when={param.kind === "domains"}
        fallback={
          <>
            <Show when={showProviderAccess}>
              <Form.Item
                extra={t("workflow.template.apply.form.provider_access.help")}
                label={t("workflow.template.apply.form.provider_access.label")}
                name={deployAccessFieldName(param.id)}
                rules={providerInfo ? [formRule] : []}
              >
                <AccessSelect
                  placeholder={t("workflow.template.apply.form.provider_access.placeholder")}
                  showSearch
                  onFilter={getAccessOptionFilter(provider!)}
                />
              </Form.Item>
            </Show>

            <ProviderConfigFields parentNamePath={deployConfigFieldName(param.id)} provider={provider ?? ""} />

            {paramErr && !isAccessRelatedParamError(paramErr) && (
              <Typography.Text type="danger" className="text-xs">
                {localizeParamError(paramErr, t)}
              </Typography.Text>
            )}
          </>
        }
      >
        <Form.Item
          extra={param.description ? void 0 : t("workflow.template.apply.form.domains.help")}
          label={t("workflow.template.apply.form.domains.label")}
          name={param.id}
          rules={[formRule]}
        >
          <MultipleSplitValueInput
            modalTitle={t("workflow.template.apply.form.domains.modal.title")}
            placeholder={t("workflow.template.apply.form.domains.placeholder")}
            placeholderInModal={t("workflow.template.apply.form.domains.modal.placeholder")}
            separator={MULTIPLE_INPUT_SEPARATOR}
            splitOptions={{ removeEmpty: true, trimSpace: true }}
          />
        </Form.Item>
      </Show>
    </div>
  );
};

type ProviderConfigFieldsProps = {
  parentNamePath: string;
  provider: string;
};

const ProviderConfigFields = ({ parentNamePath, provider }: ProviderConfigFieldsProps) => {
  const { t } = useTranslation();

  const catalogVersion = usePluginCatalogStore((s) => s.version);

  const [loaded, setLoaded] = useState<{ provider: string; envelope: ProviderSchemaEnvelope | null }>();
  useEffect(() => {
    if (!provider) return;

    let cancelled = false;
    void getProviderSchema(provider, "deploy").then((env) => {
      if (!cancelled) setLoaded({ provider: provider, envelope: env });
    });
    return () => {
      cancelled = true;
    };
  }, [provider, catalogVersion]);

  const envelope = provider === "" ? null : loaded?.provider === provider ? loaded.envelope : undefined;

  const nestedComponent = BizDeployNodeConfigFieldsProvider.useComponent(provider, {});

  let content: React.ReactNode;
  if (envelope === undefined) {
    content = (
      <div className="flex h-16 items-center justify-center">
        <Spin />
      </div>
    );
  } else if (envelope) {
    content = <SchemaConfigFields envelope={envelope} />;
  } else if (nestedComponent) {
    content = nestedComponent;
  } else {
    content = <Alert showIcon title={t("workflow.template.apply.text.config_unavailable")} type="warning" />;
  }

  return <FormNestedFieldsContextProvider value={{ parentNamePath }}>{content}</FormNestedFieldsContextProvider>;
};

export default WorkflowTemplateApplyDrawer;
