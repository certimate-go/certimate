import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { IconDots, IconEdit, IconPlayerPlay, IconSearch, IconTrash } from "@tabler/icons-react";
import { useMount } from "ahooks";
import { App, Button, Card, Dropdown, Empty, Input, Modal, Spin, Tag, Typography } from "antd";
import { ClientResponseError } from "pocketbase";

import { getWorkflowTemplate, listWorkflowTemplates, removeWorkflowTemplate } from "@/api/workflowtemplates";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import WorkflowTemplateApplyDrawer from "@/components/workflow/WorkflowTemplateApplyDrawer";
import WorkflowTemplateSaveModal from "@/components/workflow/WorkflowTemplateSaveModal";
import { WORKFLOW_TEMPLATE_PARAM_KINDS, type WorkflowTemplateModel, sortWorkflowTemplateParams } from "@/domain/workflowTemplate";
import { useZustandShallowSelector } from "@/hooks";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { unwrapErrMsg } from "@/utils/error";
import { matchSearchOption } from "@/utils/search";

const PresetListWorkflowTemplates = () => {
  const { t } = useTranslation();

  const navigate = useNavigate();

  const { modal, notification } = App.useApp();

  const { loaded: pluginCatalogLoaded, init: pluginCatalogInit } = usePluginCatalogStore(useZustandShallowSelector(["loaded", "init"]));
  useMount(() => {
    if (!pluginCatalogLoaded) {
      pluginCatalogInit();
    }
  });

  const [templates, setTemplates] = useState<WorkflowTemplateModel[]>();
  const [keyword, setKeyword] = useState("");

  const filteredTemplates = useMemo(() => {
    if (!keyword) return templates ?? [];
    return (templates ?? []).filter((template) => matchSearchOption(keyword, { label: template.name }));
  }, [templates, keyword]);

  const reload = async () => {
    const res = await listWorkflowTemplates();
    setTemplates(res);
    return res;
  };

  useMount(() => {
    reload().catch((err) => {
      if (err instanceof ClientResponseError && err.isAbort) {
        return;
      }

      console.error(err);
      setTemplates([]);
    });
  });

  const [applyDrawerTemplate, setApplyDrawerTemplate] = useState<WorkflowTemplateModel>();
  const [editTemplateId, setEditTemplateId] = useState<string>();
  const [previewTemplate, setPreviewTemplate] = useState<WorkflowTemplateModel>();

  const handleDeleteClick = (template: WorkflowTemplateModel) => {
    modal.confirm({
      title: <span className="text-error">{t("preset.action.delete.modal.title", { name: template.name })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("preset.action.delete.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await removeWorkflowTemplate(template.id);
          await reload();
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  const handleUseClick = (template: WorkflowTemplateModel) => {
    getWorkflowTemplate(template.id)
      .then((detail) => setApplyDrawerTemplate(detail))
      .catch((err) => {
        console.error(err);
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
      });
  };

  return (
    <>
      <Tips className="mb-4" message={<span dangerouslySetInnerHTML={{ __html: t("preset.props.usage.workflow.tips") }}></span>} />

      <div className="mb-4 flex items-center gap-4">
        <Input
          className="flex-1"
          allowClear
          suffix={<IconSearch size="1.25em" />}
          placeholder={t("preset.workflow.search.placeholder")}
          onChange={(e) => setKeyword(e.target.value.trim())}
        />
      </div>

      <Show
        when={(filteredTemplates?.length ?? 0) > 0}
        fallback={
          <Show
            when={templates == null}
            fallback={
              <div className="py-12">
                <Empty description={t("preset.workflow.empty")} />
              </div>
            }
          >
            <div className="flex h-40 items-center justify-center">
              <Spin />
            </div>
          </Show>
        }
      >
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-3">
          {filteredTemplates!.map((template) => {
            const params = sortWorkflowTemplateParams(template.params ?? []);
            const domainsCount = params.filter((p) => p.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DOMAINS).length;
            const deployCount = params.filter((p) => p.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DEPLOYMENT).length;

            return (
              <Card key={template.id} className="group/card size-full" hoverable onClick={() => handleUseClick(template)}>
                <div className="flex items-start justify-between gap-2 overflow-hidden">
                  <div className="min-w-0 flex-1">
                    <div className="truncate font-medium transition-colors group-hover/card:text-primary">{template.name}</div>
                    <Typography.Paragraph className="my-1 min-h-8" type="secondary" ellipsis={{ rows: 2 }}>
                      {template.description || " "}
                    </Typography.Paragraph>
                    <div className="flex items-center gap-1.5 text-xs text-gray-500">
                      {domainsCount > 0 && (
                        <Tag className="ms-0" color="default">
                          {t("preset.workflow.tag.domains", { count: domainsCount })}
                        </Tag>
                      )}
                      {deployCount > 0 && (
                        <Tag className="ms-0" color="default">
                          {t("preset.workflow.tag.deployments", { count: deployCount })}
                        </Tag>
                      )}
                    </div>
                  </div>
                  <div
                    className="shrink-0"
                    onClick={(e) => {
                      e.stopPropagation();
                    }}
                  >
                    <Dropdown
                      menu={{
                        items: [
                          {
                            key: "use",
                            label: t("preset.workflow.menu.use"),
                            icon: <IconPlayerPlay size="1.25em" />,
                            onClick: () => handleUseClick(template),
                          },
                          {
                            key: "edit",
                            label: t("preset.workflow.menu.edit"),
                            icon: <IconEdit size="1.25em" />,
                            onClick: () => setEditTemplateId(template.id),
                          },
                          {
                            key: "preview",
                            label: t("preset.workflow.menu.preview"),
                            onClick: () => setPreviewTemplate(template),
                          },
                          { type: "divider" },
                          {
                            key: "delete",
                            label: <span className="text-error">{t("preset.workflow.menu.delete")}</span>,
                            icon: <IconTrash className="text-error" size="1.25em" />,
                            danger: true,
                            onClick: () => handleDeleteClick(template),
                          },
                        ],
                      }}
                      trigger={["click"]}
                    >
                      <Button icon={<IconDots size="1.25em" />} type="text" />
                    </Dropdown>
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      </Show>

      <WorkflowTemplateApplyDrawer
        open={!!applyDrawerTemplate}
        template={applyDrawerTemplate}
        onInstantiated={(workflowId) => navigate(`/workflows/${workflowId}`, { replace: true })}
        onOpenChange={(open) => {
          if (!open) setApplyDrawerTemplate(undefined);
        }}
      />

      <WorkflowTemplatePreviewModal template={previewTemplate} onOpenChange={(open) => !open && setPreviewTemplate(undefined)} />

      <WorkflowTemplateSaveModal
        initialTemplateId={editTemplateId}
        mode="update"
        open={!!editTemplateId}
        onOpenChange={(open) => {
          if (!open) setEditTemplateId(undefined);
        }}
        onSaved={() => reload()}
      />
    </>
  );
};

const WorkflowTemplatePreviewModal = ({ template, onOpenChange }: { template?: WorkflowTemplateModel; onOpenChange: (open: boolean) => void }) => {
  const { t } = useTranslation();

  const params = useMemo(() => (template ? sortWorkflowTemplateParams(template.params) : []), [template]);

  const renderNodes = (nodes: WorkflowTemplateModel["graph"]["nodes"], depth = 0): React.ReactNode => {
    return (nodes ?? []).map((node) => (
      <div key={node.id} style={{ paddingLeft: depth * 16 }}>
        <div className="flex items-center gap-2 py-0.5">
          <Tag className="ms-0" color="default">
            {node.type}
          </Tag>
          <Typography.Text className="text-xs">{node.data?.name}</Typography.Text>
        </div>
        {(node.blocks?.length ?? 0) > 0 && renderNodes(node.blocks!, depth + 1)}
      </div>
    ));
  };

  return (
    <Modal
      closable
      destroyOnHidden
      footer={
        <Button type="primary" onClick={() => onOpenChange(false)}>
          {t("common.button.close")}
        </Button>
      }
      open={!!template}
      title={template ? t("preset.workflow.preview.title", { name: template.name }) : ""}
      width="560px"
      onCancel={() => onOpenChange(false)}
    >
      {template && (
        <div className="pt-2">
          <Typography.Title className="mt-0!" level={5}>
            {t("preset.workflow.preview.params")}
          </Typography.Title>
          {params.length === 0 && <Typography.Text type="secondary">{t("preset.workflow.preview.no_params")}</Typography.Text>}
          {params.map((param) => (
            <div key={param.id} className="mb-1 flex items-baseline gap-2">
              <Tag color="default">
                {param.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DOMAINS ? t("workflow.template.apply.tag.domains") : t("workflow.template.apply.tag.deployment")}
              </Tag>
              <span className="text-sm">{param.title}</span>
              {param.description && (
                <Typography.Text className="text-xs" type="secondary">
                  {param.description}
                </Typography.Text>
              )}
            </div>
          ))}

          <Typography.Title level={5}>{t("preset.workflow.preview.graph")}</Typography.Title>
          <div className="rounded-lg border border-gray-200 p-2 dark:border-gray-700">{renderNodes(template.graph?.nodes)}</div>
        </div>
      )}
    </Modal>
  );
};

export default PresetListWorkflowTemplates;
