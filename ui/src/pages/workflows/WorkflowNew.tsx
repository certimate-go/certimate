import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { IconArrowRight, IconSquarePlus2, IconUpload } from "@tabler/icons-react";
import { App, Button, Card, Spin, Typography } from "antd";

import Show from "@/components/Show";
import WorkflowGraphImportModal from "@/components/workflow/WorkflowGraphImportModal";
import {
  WORKFLOW_NODE_TYPES,
  type WorkflowModel,
  type WorkflowNodeConfigForBizDeploy,
  type WorkflowNodeConfigForBizNotify,
  type WorkflowNodeConfigForBranchBlock,
  newNode,
} from "@/domain/workflow";
import { save as saveWorkflow } from "@/repository/workflow";
import { getErrMsg } from "@/utils/error";

const TEMPLATE_KEY_BLANK = "blank" as const;
const TEMPLATE_KEY_STANDARD = "standard" as const;
const TEMPLATE_KEY_CERTTEST = "certtest" as const;
type TemplateKeys = typeof TEMPLATE_KEY_BLANK | typeof TEMPLATE_KEY_CERTTEST | typeof TEMPLATE_KEY_STANDARD;

const WorkflowNew = () => {
  const navigate = useNavigate();

  const { i18n, t } = useTranslation();

  const { notification } = App.useApp();

  const templates = [
    {
      key: TEMPLATE_KEY_STANDARD,
      name: t("workflow.new.templates.template.standard.title"),
      description: t("workflow.new.templates.template.standard.description"),
      image: "/imgs/workflow/tpl-standard.png",
    },
    {
      key: TEMPLATE_KEY_CERTTEST,
      name: t("workflow.new.templates.template.certtest.title"),
      description: t("workflow.new.templates.template.certtest.description"),
      image: "/imgs/workflow/tpl-certtest.png",
    },
  ];
  const [templateSelectKey, setTemplateSelectKey] = useState<TemplateKeys>();
  const [templatePending, setTemplatePending] = useState(false);

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
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });

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
        notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      } finally {
        setTemplatePending(false);
      }
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("workflow.new.title")}</h1>
        <p className="text-base text-gray-500">{t("workflow.new.subtitle")}</p>
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
          <h3>{t("workflow.new.templates.title")}</h3>
          <Typography.Text type="secondary">
            <div className="mb-4">{t("workflow.new.templates.subtitle")}</div>
          </Typography.Text>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
            {templates.map((template) => renderTemplateCard(template))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkflowNew;
