import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useControllableValue, useMount } from "ahooks";
import { Alert, App, Button, Checkbox, Divider, Form, type FormInstance, Input, Modal, Select, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { ClientResponseError } from "pocketbase";
import { z } from "zod";

import { createWorkflowTemplate, getWorkflowTemplate, listWorkflowTemplates, updateWorkflowTemplate } from "@/api/workflowtemplates";
import Show from "@/components/Show";
import { WORKFLOW_NODE_TYPES, type WorkflowGraph, type WorkflowNode } from "@/domain/workflow";
import {
  WORKFLOW_TEMPLATE_PARAM_KINDS,
  type WorkflowTemplateModel,
  type WorkflowTemplateParam,
  findWorkflowGraphNodeById,
  listWorkflowGraphNodesByType,
  sortWorkflowTemplateParams,
} from "@/domain/workflowTemplate";
import { useAntdForm } from "@/hooks";
import { unwrapErrMsg } from "@/utils/error";

export interface WorkflowTemplateSaveModalProps {
  afterClose?: () => void;
  getGraph?: () => WorkflowGraph | undefined;
  initialTemplateId?: string;
  mode?: "create" | "update";
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSaved?: (template: WorkflowTemplateModel) => void;
}

const DOMAINS_PARAM_ID = "domains";
const deployParamId = (nodeId: string) => `deploy_${nodeId}`;
const paramIdOfDeployNode = (param: WorkflowTemplateParam) => param.deployNodeId ?? "";

type BoundState = {
  applyNodeIds: string[];
  deployNodeIds: string[];
  domainsTitle: string;
  domainsDescription: string;
  deployTitles: Record<string, string>;
};

const WorkflowTemplateSaveModal = ({ afterClose, getGraph, initialTemplateId, mode = "create", ...props }: WorkflowTemplateSaveModalProps) => {
  const { t } = useTranslation();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const [formInst] = Form.useForm();

  return (
    <Modal
      afterClose={afterClose}
      closable
      destroyOnHidden
      footer={
        <div className="flex items-center justify-between gap-4">
          <Typography.Text type="secondary" className="text-xs">
            {t("workflow.template.save.text.footer_notice")}
          </Typography.Text>
          <div className="flex gap-2">
            <Button onClick={() => setOpen(false)}>{t("common.button.cancel")}</Button>
            <Button type="primary" onClick={() => formInst.submit()}>
              {mode === "update" ? t("workflow.template.save.button.update") : t("workflow.template.save.button.save")}
            </Button>
          </div>
        </div>
      }
      open={open}
      title={mode === "update" ? t("workflow.template.save.title_update") : t("workflow.template.save.title")}
      width="896px"
      onCancel={() => setOpen(false)}
    >
      <SaveModalContent
        formInst={formInst}
        getGraph={getGraph}
        initialTemplateId={initialTemplateId}
        mode={mode}
        onOpenChange={setOpen}
        onSaved={props.onSaved}
      />
    </Modal>
  );
};

type SaveModalContentProps = {
  formInst: FormInstance;
  getGraph?: () => WorkflowGraph | undefined;
  initialTemplateId?: string;
  mode: "create" | "update";
  onOpenChange: (open: boolean) => void;
  onSaved?: (template: WorkflowTemplateModel) => void;
};

const SaveModalContent = ({ formInst, getGraph, initialTemplateId, mode, onOpenChange, onSaved }: SaveModalContentProps) => {
  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const [canvasGraph] = useState<WorkflowGraph | undefined>(() => getGraph?.());

  const [templates, setTemplates] = useState<WorkflowTemplateModel[]>([]);
  const [targetTemplate, setTargetTemplate] = useState<WorkflowTemplateModel>();
  const [targetPending, setTargetPending] = useState(false);
  const [revisionConflict, setRevisionConflict] = useState(false);

  const effectiveGraph = canvasGraph ?? targetTemplate?.graph;

  const applyNodes = useMemo(
    () =>
      listWorkflowGraphNodesByType(effectiveGraph, WORKFLOW_NODE_TYPES.BIZ_APPLY).filter(
        (node) => (node.data?.config as Record<string, unknown>)?.identifier === "domain"
      ),
    [effectiveGraph]
  );
  const deployNodes = useMemo(() => listWorkflowGraphNodesByType(effectiveGraph, WORKFLOW_NODE_TYPES.BIZ_DEPLOY), [effectiveGraph]);

  const [bound, setBound] = useState<BoundState>(() => ({
    applyNodeIds: mode === "create" ? applyNodes.map((n) => n.id) : [],
    deployNodeIds: mode === "create" ? deployNodes.map((n) => n.id) : [],
    domainsTitle: mode === "create" ? t("workflow.template.save.domains_title.default") : "",
    domainsDescription: "",
    deployTitles: mode === "create" ? Object.fromEntries(deployNodes.map((n) => [n.id, n.data?.name ?? ""])) : {},
  }));

  const formSchema = z.object({
    targetTemplateId: mode === "update" && !initialTemplateId ? z.string().nonempty() : z.string().nullish(),
    name: z.string().nonempty(t("workflow.template.save.form.name.errmsg.invalid")),
    description: z.string().nullish(),
    domainsTitle: z
      .string()
      .nullish()
      .refine((v) => bound.applyNodeIds.length === 0 || !!v?.trim(), t("workflow.template.save.form.domains_title.errmsg.invalid")),
    domainsDescription: z.string().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const { formProps } = useAntdForm({
    form: formInst,
    name: "workflowTemplateSaveForm",
    initialValues: {
      targetTemplateId: initialTemplateId ?? "",
      name: "",
      description: "",
      domainsTitle: mode === "create" ? t("workflow.template.save.domains_title.default") : "",
      domainsDescription: "",
    },
    onSubmit: handleSubmit,
  });

  useMount(() => {
    if (mode !== "update" || initialTemplateId) return;

    listWorkflowTemplates().then((res) => {
      setTemplates(res);
    });
  });

  const applyBinding = (template: WorkflowTemplateModel) => {
    const params = sortWorkflowTemplateParams(template.params ?? []);
    const domainsParam = params.find((p) => p.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DOMAINS);
    const templateApplyNodes = listWorkflowGraphNodesByType(template.graph, WORKFLOW_NODE_TYPES.BIZ_APPLY).filter(
      (node) => (node.data?.config as Record<string, unknown>)?.identifier === "domain"
    );
    const templateDeployNodes = listWorkflowGraphNodesByType(template.graph, WORKFLOW_NODE_TYPES.BIZ_DEPLOY);
    const applyNodeIds = (domainsParam?.applyNodeIds ?? []).filter((id) => templateApplyNodes.some((n) => n.id === id));
    const deployNodeIds = params
      .filter((p) => p.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DEPLOYMENT)
      .map((p) => paramIdOfDeployNode(p))
      .filter((id) => templateDeployNodes.some((n) => n.id === id));
    const deployTitles: Record<string, string> = {};
    for (const p of params) {
      if (p.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DEPLOYMENT && p.deployNodeId) deployTitles[p.deployNodeId] = p.title;
    }

    setBound({
      applyNodeIds,
      deployNodeIds,
      domainsTitle: applyNodeIds.length > 0 ? (domainsParam?.title ?? "") : "",
      domainsDescription: domainsParam?.description ?? "",
      deployTitles,
    });
    formInst.setFieldsValue({
      name: template.name,
      description: template.description ?? "",
      domainsTitle: applyNodeIds.length > 0 ? (domainsParam?.title ?? "") : "",
      domainsDescription: domainsParam?.description ?? "",
    });
  };

  const handleTargetTemplateChange = async (templateId: string) => {
    setRevisionConflict(false);
    setTargetPending(true);
    try {
      const template = await getWorkflowTemplate(templateId);
      setTargetTemplate(template);
      applyBinding(template);
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    } finally {
      setTargetPending(false);
    }
  };

  useEffect(() => {
    if (mode !== "update" || !initialTemplateId) return;

    let cancelled = false;
    getWorkflowTemplate(initialTemplateId)
      .then((template) => {
        if (cancelled) return;
        setTargetTemplate(template);
        applyBinding(template);
      })
      .catch((err) => {
        console.error(err);
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const toggleApplyNode = (nodeId: string, checked: boolean) => {
    setBound((prev) => ({
      ...prev,
      applyNodeIds: checked ? [...prev.applyNodeIds, nodeId] : prev.applyNodeIds.filter((id) => id !== nodeId),
      domainsTitle: checked && prev.applyNodeIds.length === 0 ? t("workflow.template.save.domains_title.default") : prev.domainsTitle,
    }));
  };

  const toggleDeployNode = (nodeId: string, nodeName: string, checked: boolean) => {
    setBound((prev) => ({
      ...prev,
      deployNodeIds: checked ? [...prev.deployNodeIds, nodeId] : prev.deployNodeIds.filter((id) => id !== nodeId),
      deployTitles: checked ? { ...prev.deployTitles, [nodeId]: prev.deployTitles[nodeId] || nodeName } : prev.deployTitles,
    }));
  };

  async function handleSubmit(values: Record<string, unknown>) {
    if (!effectiveGraph) return;

    const params: WorkflowTemplateParam[] = [];
    let order = 0;
    if (bound.applyNodeIds.length > 0) {
      params.push({
        id: DOMAINS_PARAM_ID,
        kind: WORKFLOW_TEMPLATE_PARAM_KINDS.DOMAINS,
        title: String(values.domainsTitle ?? "").trim(),
        description: String(values.domainsDescription ?? "").trim(),
        order: order++,
        applyNodeIds: [...bound.applyNodeIds],
      });
    }
    for (const nodeId of bound.deployNodeIds) {
      params.push({
        id: deployParamId(nodeId),
        kind: WORKFLOW_TEMPLATE_PARAM_KINDS.DEPLOYMENT,
        title: (bound.deployTitles[nodeId] ?? "").trim(),
        order: order++,
        deployNodeId: nodeId,
      });
    }

    const req = {
      name: String(values.name ?? "").trim(),
      description: String(values.description ?? "").trim(),
      graph: effectiveGraph,
      params: params,
    };

    try {
      if (mode === "update") {
        if (!targetTemplate) return;

        const template = await updateWorkflowTemplate(targetTemplate.id, { ...req, expectedRevision: targetTemplate.revision });
        message.success(t("workflow.template.save.text.updated"));
        onOpenChange(false);
        onSaved?.(template);
      } else {
        const template = await createWorkflowTemplate(req);
        message.success(t("workflow.template.save.text.saved"));
        onOpenChange(false);
        onSaved?.(template);
      }
    } catch (err) {
      if (err instanceof ClientResponseError && err.status === 409) {
        setRevisionConflict(true);
        if (targetTemplate) {
          try {
            const latest = await getWorkflowTemplate(targetTemplate.id);
            setTargetTemplate(latest);
          } catch (refreshErr) {
            console.error(refreshErr);
          }
        }
        return;
      }

      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    }
  }

  return (
    <div className="pt-2">
      {revisionConflict && (
        <Alert
          className="mb-3"
          showIcon
          title={t("workflow.template.save.revision_conflict.title")}
          description={t("workflow.template.save.revision_conflict.description")}
          type="warning"
        />
      )}

      <Form {...formProps} clearOnDestroy={false} form={formInst} layout="vertical" preserve={true}>
        <Show when={mode === "update" && !initialTemplateId}>
          <Form.Item label={t("workflow.template.save.form.target.label")} name="targetTemplateId" rules={[formRule]}>
            <Select
              loading={targetPending}
              options={templates.map((tpl) => ({ label: tpl.name, value: tpl.id }))}
              placeholder={t("workflow.template.save.form.target.placeholder")}
              showSearch
              optionFilterProp="label"
              onSelect={handleTargetTemplateChange}
            />
          </Form.Item>
        </Show>

        <div className="grid grid-cols-1 gap-x-6 md:grid-cols-2">
          <div>
            <Form.Item className="mb-2" label={t("workflow.template.save.form.name.label")} name="name" rules={[formRule]}>
              <Input placeholder={t("workflow.template.save.form.name.placeholder")} />
            </Form.Item>
            <Form.Item className="mb-2" label={t("workflow.template.save.form.description.label")} name="description">
              <Input placeholder={t("workflow.template.save.form.description.placeholder")} />
            </Form.Item>

            <Divider className="mt-2 mb-1" size="small" titlePlacement="left">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow.template.save.section.apply_nodes")}
              </Typography.Text>
            </Divider>
            {applyNodes.map((node) => (
              <ApplyNodeRow
                key={node.id}
                node={node}
                checked={bound.applyNodeIds.includes(node.id)}
                onToggle={(checked) => toggleApplyNode(node.id, checked)}
              />
            ))}
            {applyNodes.length === 0 && (
              <Typography.Text type="secondary" className="text-xs">
                {t("workflow.template.save.text.no_apply_nodes")}
              </Typography.Text>
            )}

            {bound.applyNodeIds.length > 0 && (
              <>
                <Form.Item className="mt-2 mb-1" label={t("workflow.template.save.form.domains_title.label")} name="domainsTitle" rules={[formRule]}>
                  <Input placeholder={t("workflow.template.save.form.domains_title.placeholder")} />
                </Form.Item>
                <Form.Item className="mb-1" label={t("workflow.template.save.form.domains_description.label")} name="domainsDescription">
                  <Input placeholder={t("workflow.template.save.form.domains_description.placeholder")} />
                </Form.Item>
              </>
            )}

            <Divider className="mt-3 mb-1" size="small" titlePlacement="left">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow.template.save.section.deploy_nodes")}
              </Typography.Text>
            </Divider>
            {deployNodes.map((node) => (
              <DeployNodeRow
                key={node.id}
                node={node}
                checked={bound.deployNodeIds.includes(node.id)}
                title={bound.deployTitles[node.id] ?? ""}
                onTitleChange={(title) => setBound((prev) => ({ ...prev, deployTitles: { ...prev.deployTitles, [node.id]: title } }))}
                onToggle={(checked) => toggleDeployNode(node.id, node.data?.name ?? "", checked)}
              />
            ))}
            {deployNodes.length === 0 && (
              <Typography.Text type="secondary" className="text-xs">
                {t("workflow.template.save.text.no_deploy_nodes")}
              </Typography.Text>
            )}
          </div>

          <div className="rounded-lg border border-gray-200 p-3 dark:border-gray-700">
            <div className="mb-2 text-xs font-semibold text-gray-500">{t("workflow.template.save.preview.title")}</div>
            <Show when={bound.applyNodeIds.length > 0}>
              <div className="mb-3">
                <div className="text-sm font-medium">{bound.domainsTitle || t("workflow.template.apply.form.domains.label")}</div>
                <div className="mt-1 rounded-sm border border-dashed border-gray-300 px-2 py-1.5 text-xs text-gray-400 dark:border-gray-600">
                  {t("workflow.template.apply.form.domains.placeholder")}
                </div>
                <div className="mt-1 text-xs text-gray-400">{t("workflow.template.save.preview.apply_nodes", { count: bound.applyNodeIds.length })}</div>
              </div>
            </Show>
            {bound.deployNodeIds.map((nodeId) => {
              const node = findWorkflowGraphNodeById(effectiveGraph, nodeId);
              const provider = ((node?.data?.config ?? {}) as Record<string, unknown>).provider as string | undefined;
              return (
                <div key={nodeId} className="mb-3">
                  <div className="text-sm font-medium">{bound.deployTitles[nodeId]}</div>
                  <div className="mt-1 rounded-sm border border-dashed border-gray-300 px-2 py-1.5 text-xs text-gray-400 dark:border-gray-600">
                    {t("workflow.template.apply.form.provider_access.placeholder")}
                  </div>
                  <div className="mt-1 text-xs text-gray-400">{t("workflow.template.save.preview.deploy_provider", { provider: provider ?? "" })}</div>
                </div>
              );
            })}
            <Show when={bound.applyNodeIds.length === 0 && bound.deployNodeIds.length === 0}>
              <div className="text-xs text-gray-400">{t("workflow.template.save.preview.empty")}</div>
            </Show>
          </div>
        </div>
      </Form>
    </div>
  );
};

type NodeRowProps = {
  checked: boolean;
  node: WorkflowNode;
  onToggle: (checked: boolean) => void;
};

const ApplyNodeRow = ({ checked, node, onToggle }: NodeRowProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex items-center gap-2 py-1">
      <Checkbox checked={checked} onChange={(e) => onToggle(e.target.checked)}>
        <span className="text-sm">{node.data?.name || node.id}</span>
      </Checkbox>
      <Typography.Text type="secondary" className="text-xs">
        {t("workflow.template.save.text.expose_domains")}
      </Typography.Text>
    </div>
  );
};

type DeployNodeRowProps = {
  checked: boolean;
  node: WorkflowNode;
  title: string;
  onTitleChange: (title: string) => void;
  onToggle: (checked: boolean) => void;
};

const DeployNodeRow = ({ checked, node, title, onTitleChange, onToggle }: DeployNodeRowProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex items-center gap-2 py-1">
      <Checkbox checked={checked} onChange={(e) => onToggle(e.target.checked)}>
        <span className="text-sm">{node.data?.name || node.id}</span>
      </Checkbox>
      {checked && (
        <Input
          className="flex-1"
          size="small"
          value={title}
          placeholder={t("workflow.template.save.form.deploy_title.placeholder")}
          onChange={(e) => onTitleChange(e.target.value)}
        />
      )}
    </div>
  );
};

export default WorkflowTemplateSaveModal;
