import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { IconArrowRight, IconSquarePlus2, IconUpload } from "@tabler/icons-react";
import { useMount } from "ahooks";
import { App, Button, Card, Spin, Tag, Typography } from "antd";
import { ClientResponseError } from "pocketbase";

import { getWorkflowTemplate, listWorkflowTemplates } from "@/api/workflowtemplates";
import Show from "@/components/Show";
import WorkflowGraphImportModal from "@/components/workflow/WorkflowGraphImportModal";
import WorkflowTemplateApplyDrawer from "@/components/workflow/WorkflowTemplateApplyDrawer";
import {
  WORKFLOW_NODE_TYPES,
  type WorkflowModel,
  type WorkflowNodeConfigForBizDeploy,
  type WorkflowNodeConfigForBizNotify,
  type WorkflowNodeConfigForBranchBlock,
  newNode,
} from "@/domain/workflow";
import { WORKFLOW_TEMPLATE_PARAM_KINDS, type WorkflowTemplateModel, sortWorkflowTemplateParams } from "@/domain/workflowTemplate";
import { save as saveWorkflow } from "@/repository/workflow";
import { unwrapErrMsg } from "@/utils/error";
import { withBasePath } from "@/utils/url";

const TEMPLATE_KEY_BLANK = "blank" as const;
const TEMPLATE_KEY_STANDARD = "standard" as const;
const TEMPLATE_KEY_CERTTEST = "certtest" as const;
type TemplateKeys = typeof TEMPLATE_KEY_BLANK | typeof TEMPLATE_KEY_CERTTEST | typeof TEMPLATE_KEY_STANDARD;

const WorkflowNew = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const { i18n, t } = useTranslation();

  const { notification } = App.useApp();

  const templates = [
    {
      key: TEMPLATE_KEY_STANDARD,
      name: t("workflow.new.templates.template.standard.title"),
      description: t("workflow.new.templates.template.standard.description"),
      image: withBasePath("/imgs/workflow/tpl-standard.png"),
    },
    {
      key: TEMPLATE_KEY_CERTTEST,
      name: t("workflow.new.templates.template.certtest.title"),
      description: t("workflow.new.templates.template.certtest.description"),
      image: withBasePath("/imgs/workflow/tpl-certtest.png"),
    },
  ];
  const [templateSelectKey, setTemplateSelectKey] = useState<TemplateKeys>();
  const [templatePending, setTemplatePending] = useState(false);

  const [customTemplates, setCustomTemplates] = useState<WorkflowTemplateModel[]>();
  const [applyDrawerTemplate, setApplyDrawerTemplate] = useState<WorkflowTemplateModel>();

  useMount(() => {
    listWorkflowTemplates()
      .then((res) => {
        setCustomTemplates(res);

        const templateId = searchParams.get("template");
        if (templateId) {
          setSearchParams((prev) => {
            prev.delete("template");
            return prev;
          });
          getWorkflowTemplate(templateId).then(setApplyDrawerTemplate);
        }
      })
      .catch((err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);
        setCustomTemplates([]);
      });
  });

  const renderTemplateCard = ({ key, name, description, image }: { key: TemplateKeys; name: string; description: string; image: string }) => {
    return (
      <Card
        key={key}
        className="group/card size-full"
        cover={<img className="min-h-[120px] object-contain" src={image} />}
        hoverable
        onClick={() => handleTemplateClick(key)}
      >
        <div className="flex w-full items-center gap-4">
          <Card.Meta
            className="grow"
            title={
              <div className="flex w-full items-center justify-between gap-4 overflow-hidden transition-colors group-hover/card:text-primary">
                <div className="flex-1 truncate">{name}</div>
                <Show when={templatePending} fallback={<IconArrowRight className="opacity-0 transition-opacity group-hover/card:opacity-100" size="1.25em" />}>
                  <Spin spinning={templateSelectKey === key} />
                </Show>
              </div>
            }
            description={description}
          />
        </div>
      </Card>
    );
  };

  const renderCustomTemplateCard = (template: WorkflowTemplateModel) => {
    const params = sortWorkflowTemplateParams(template.params ?? []);
    const summary = params
      .map((param) =>
        param.kind === WORKFLOW_TEMPLATE_PARAM_KINDS.DOMAINS ? t("workflow.template.apply.tag.domains") : t("workflow.template.apply.tag.deployment")
      )
      .join(" · ");

    return (
      <Card key={template.id} className="group/card size-full" hoverable onClick={() => setApplyDrawerTemplate(template)}>
        <div className="flex w-full items-center gap-4">
          <Card.Meta
            className="grow"
            title={
              <div className="flex w-full items-center justify-between gap-4 overflow-hidden transition-colors group-hover/card:text-primary">
                <div className="flex-1 truncate">{template.name}</div>
                <IconArrowRight className="opacity-0 transition-opacity group-hover/card:opacity-100" size="1.25em" />
              </div>
            }
            description={
              <div className="flex flex-col gap-1">
                <div className="line-clamp-2 min-h-8">{template.description}</div>
                {summary && (
                  <Tag className="w-fit" color="default">
                    {summary}
                  </Tag>
                )}
              </div>
            }
          />
        </div>
      </Card>
    );
  };

  const { modalProps: workflowImportModalProps, ...workflowImportModal } = WorkflowGraphImportModal.useModal();

  const handleTemplateClick = async (key: TemplateKeys) => {
    if (templatePending) return;

    setTemplateSelectKey(key);
    setTemplatePending(true);

    try {
      let workflow = {} as WorkflowModel;
      workflow.name = t("workflow.new.templates.default_name");
      workflow.description = t("workflow.new.templates.default_description");
      workflow.graphDraft = { nodes: [] };
      workflow.hasDraft = true;

      switch (key) {
        case TEMPLATE_KEY_BLANK:
          {
            const startNode = newNode(WORKFLOW_NODE_TYPES.START, { i18n: i18n });
            const endNode = newNode(WORKFLOW_NODE_TYPES.END, { i18n: i18n });

            workflow.graphDraft!.nodes = [startNode, endNode];
          }
          break;

        case TEMPLATE_KEY_STANDARD:
          {
            const startNode = newNode(WORKFLOW_NODE_TYPES.START, { i18n: i18n });
            const tryCatchNode = newNode(WORKFLOW_NODE_TYPES.TRYCATCH, { i18n: i18n });
            const applyNode = newNode(WORKFLOW_NODE_TYPES.BIZ_APPLY, { i18n: i18n });
            const deployNode = newNode(WORKFLOW_NODE_TYPES.BIZ_DEPLOY, { i18n: i18n });
            const notifyOnFailureNode = newNode(WORKFLOW_NODE_TYPES.BIZ_NOTIFY, { i18n: i18n });
            const endNode = newNode(WORKFLOW_NODE_TYPES.END, { i18n: i18n });

            deployNode.data.config = {
              ...deployNode.data.config,
              certificateOutputNodeId: applyNode.id,
            } as WorkflowNodeConfigForBizDeploy;

            notifyOnFailureNode.data.config = {
              ...notifyOnFailureNode.data.config,
              subject: "[Certimate] Workflow Failure Alert!",
              message: 'Your workflow "{{ $workflow.name }}" run has failed. Please check the details.',
            } as WorkflowNodeConfigForBizNotify;

            tryCatchNode.blocks!.at(0)!.blocks ??= [];
            tryCatchNode.blocks!.at(0)!.blocks!.push(applyNode, deployNode);
            tryCatchNode.blocks!.at(1)!.blocks ??= [];
            tryCatchNode.blocks!.at(1)!.blocks!.unshift(notifyOnFailureNode);

            workflow.graphDraft!.nodes = [startNode, tryCatchNode, endNode];
          }
          break;

        case TEMPLATE_KEY_CERTTEST:
          {
            const startNode = newNode(WORKFLOW_NODE_TYPES.START, { i18n: i18n });
            const tryCatchNode = newNode(WORKFLOW_NODE_TYPES.TRYCATCH, { i18n: i18n });
            const monitorNode = newNode(WORKFLOW_NODE_TYPES.BIZ_MONITOR, { i18n: i18n });
            const conditionNode = newNode(WORKFLOW_NODE_TYPES.CONDITION, { i18n: i18n });
            const notifyOnExpiringSoonNode = newNode(WORKFLOW_NODE_TYPES.BIZ_NOTIFY, { i18n: i18n });
            const notifyOnExpiredNode = newNode(WORKFLOW_NODE_TYPES.BIZ_NOTIFY, { i18n: i18n });
            const notifyOnFailureNode = newNode(WORKFLOW_NODE_TYPES.BIZ_NOTIFY, { i18n: i18n });
            const endNode = newNode(WORKFLOW_NODE_TYPES.END, { i18n: i18n });

            notifyOnExpiringSoonNode.data.config = {
              ...notifyOnExpiringSoonNode.data.config,
              subject: "[Certimate] Certificate Expiry Alert!",
              message:
                "The certificate which you are monitoring will be expiring soon. Please pay attention to your website. \r\nDomains: {{ $certificate.subjectAltNames }} \r\nExpiration: {{ $certificate.notAfter }}({{ $certificate.daysLeft }} days left)",
            } as WorkflowNodeConfigForBizNotify;

            notifyOnExpiredNode.data.config = {
              ...notifyOnExpiredNode.data.config,
              subject: "[Certimate] Certificate Expiry Alert!",
              message:
                "The certificate which you are monitoring has already expired. Please pay attention to your website. \r\nDomains: {{ $certificate.subjectAltNames }} \r\nExpiration: {{ $certificate.notAfter }}",
            } as WorkflowNodeConfigForBizNotify;

            notifyOnFailureNode.data.config = {
              ...notifyOnFailureNode.data.config,
              subject: "[Certimate] Workflow Failure Alert!",
              message: 'Your workflow "{{ $workflow.name }}" run has failed. Please check the details.',
            } as WorkflowNodeConfigForBizNotify;

            tryCatchNode.blocks!.at(0)!.blocks ??= [];
            tryCatchNode.blocks!.at(0)!.blocks!.push(monitorNode, conditionNode);
            tryCatchNode.blocks!.at(1)!.blocks ??= [];
            tryCatchNode.blocks!.at(1)!.blocks!.unshift(notifyOnFailureNode);

            conditionNode.blocks!.at(0)!.data.name = t("workflow_node.condition.default_name.template_certtest_on_expiring_soon");
            conditionNode.blocks!.at(0)!.data.config = {
              ...conditionNode.blocks!.at(0)!.data.config,
              expression: {
                left: {
                  left: {
                    selector: {
                      id: monitorNode.id,
                      name: "certificate.validity",
                      type: "boolean",
                    },
                    type: "var",
                  },
                  operator: "eq",
                  right: {
                    type: "const",
                    value: "true",
                    valueType: "boolean",
                  },
                  type: "comparison",
                },
                operator: "and",
                right: {
                  left: {
                    selector: {
                      id: monitorNode.id,
                      name: "certificate.daysLeft",
                      type: "number",
                    },
                    type: "var",
                  },
                  operator: "lte",
                  right: {
                    type: "const",
                    value: "30",
                    valueType: "number",
                  },
                  type: "comparison",
                },
                type: "logical",
              },
            } as WorkflowNodeConfigForBranchBlock;
            conditionNode.blocks!.at(0)!.blocks ??= [];
            conditionNode.blocks!.at(0)!.blocks!.push(notifyOnExpiringSoonNode);
            conditionNode.blocks!.at(1)!.data.name = t("workflow_node.condition.default_name.template_certtest_on_expired");
            conditionNode.blocks!.at(1)!.data.config = {
              ...conditionNode.blocks!.at(1)!.data.config,
              expression: {
                left: {
                  selector: {
                    id: monitorNode.id,
                    name: "certificate.validity",
                    type: "boolean",
                  },
                  type: "var",
                },
                operator: "eq",
                right: {
                  type: "const",
                  value: "false",
                  valueType: "boolean",
                },
                type: "comparison",
              },
            } as WorkflowNodeConfigForBranchBlock;
            conditionNode.blocks!.at(1)!.blocks ??= [];
            conditionNode.blocks!.at(1)!.blocks!.push(notifyOnExpiredNode);

            workflow.graphDraft!.nodes = [startNode, tryCatchNode, endNode];
          }
          break;

        default:
          throw "Invalid value of `templateSelectKey`";
      }

      workflow = await saveWorkflow(workflow);
      navigate(`/workflows/${workflow.id}`, { replace: true });
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

      throw err;
    } finally {
      setTemplatePending(false);
      setTemplateSelectKey(void 0);
    }
  };

  const handleImportClick = async () => {
    if (templatePending) return;

    workflowImportModal.open().then(async (graph) => {
      setTemplatePending(true);

      try {
        let workflow = {} as WorkflowModel;
        workflow.name = t("workflow.new.templates.default_name");
        workflow.description = t("workflow.new.templates.default_description");
        workflow.graphDraft = graph;
        workflow.hasDraft = true;
        workflow = await saveWorkflow(workflow);
        navigate(`/workflows/${workflow.id}`, { replace: true });
      } catch (err) {
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

        throw err;
      } finally {
        setTemplatePending(false);
      }
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("workflow.new.heading")}</h1>
        <p className="text-base text-gray-500">{t("workflow.new.subheading")}</p>
      </div>

      <div className="container">
        <div className="my-1.5">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
            <Card className="size-full" styles={{ body: { padding: "1rem 0.5rem" } }} variant="borderless">
              <div className="flex flex-col gap-3">
                <Button block icon={<IconSquarePlus2 size="1.25em" />} type="text" onClick={() => handleTemplateClick(TEMPLATE_KEY_BLANK)}>
                  <div className="w-full text-left">{t("workflow.new.button.create")}</div>
                </Button>
                <Button block icon={<IconUpload size="1.25em" />} type="text" onClick={handleImportClick}>
                  <div className="w-full text-left">{t("workflow.new.button.import")}</div>
                </Button>
              </div>
            </Card>

            <WorkflowGraphImportModal {...workflowImportModalProps} />
          </div>
        </div>

        <div className="mt-8">
          <h3>{t("workflow.new.custom_templates.title")}</h3>
          <Typography.Text type="secondary">
            <div className="mb-4">{t("workflow.new.custom_templates.subtitle")}</div>
          </Typography.Text>

          <div>
            <Show
              when={(customTemplates?.length ?? 0) > 0}
              fallback={
                <Card className="size-full" styles={{ body: { padding: "1rem" } }} variant="borderless">
                  <Typography.Text type="secondary">
                    {t("workflow.new.custom_templates.empty")}
                    <Button className="ms-2 p-0" size="small" type="link" onClick={() => navigate("/presets?usage=workflow")}>
                      {t("workflow.new.custom_templates.manage")}
                    </Button>
                  </Typography.Text>
                </Card>
              }
            >
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
                {customTemplates?.map((tpl) => renderCustomTemplateCard(tpl))}
              </div>
            </Show>
          </div>
        </div>

        <div className="mt-8">
          <h3>{t("workflow.new.templates.title")}</h3>
          <Typography.Text type="secondary">
            <div className="mb-4">{t("workflow.new.templates.subtitle")}</div>
          </Typography.Text>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
            {templates.map((template) => renderTemplateCard(template))}
          </div>
        </div>
      </div>

      <WorkflowTemplateApplyDrawer
        open={!!applyDrawerTemplate}
        template={applyDrawerTemplate}
        onInstantiated={(workflowId) => navigate(`/workflows/${workflowId}`, { replace: true })}
        onOpenChange={(open) => {
          if (!open) setApplyDrawerTemplate(void 0);
        }}
      />
    </div>
  );
};

export default WorkflowNew;
