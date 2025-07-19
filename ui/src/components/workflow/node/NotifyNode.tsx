import { memo, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Avatar, Flex, Typography } from "antd";
import { produce } from "immer";

import { notificationProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForNotify, WorkflowNodeType } from "@/domain/workflow";
import { useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";

import SharedNode, { type SharedNodeProps } from "./_SharedNode";
import NotifyNodeConfigForm, { type NotifyNodeConfigFormInstance } from "./NotifyNodeConfigForm";

export type NotifyNodeProps = SharedNodeProps;

const NotifyNode = ({ node, disabled }: NotifyNodeProps) => {
  if (node.type !== WorkflowNodeType.Notify) {
    console.warn(`[certimate] current workflow node type is not: ${WorkflowNodeType.Notify}`);
  }

  const { t } = useTranslation();

  const { updateNode } = useWorkflowStore(useZustandShallowSelector(["updateNode"]));

  const formRef = useRef<NotifyNodeConfigFormInstance>(null);
  const [formPending, setFormPending] = useState(false);
  const getFormValues = () => formRef.current!.getFieldsValue() as WorkflowNodeConfigForNotify;

  const [drawerOpen, setDrawerOpen] = useState(false);

  const wrappedEl = useMemo(() => {
    if (node.type !== WorkflowNodeType.Notify) {
      console.warn(`[certimate] current workflow node type is not: ${WorkflowNodeType.Notify}`);
    }

    if (!node.validated) {
      return <Typography.Link>{t("workflow_node.action.configure_node")}</Typography.Link>;
    }

    const config = (node.config as WorkflowNodeConfigForNotify) ?? {};
    const provider = notificationProvidersMap.get(config.provider);
    return (
      <Flex className="size-full overflow-hidden" align="center" gap={8}>
        <Avatar shape="square" src={provider?.icon} size="small" />
        <Typography.Text className="flex-1 truncate">{t(provider?.name ?? "　")}</Typography.Text>
        <Typography.Text className="truncate" type="secondary">
          {config.subject ?? ""}
        </Typography.Text>
      </Flex>
    );
  }, [node]);

  const handleDrawerConfirm = async () => {
    setFormPending(true);
    try {
      await formRef.current!.validateFields();
    } catch (err) {
      setFormPending(false);
      throw err;
    }

    try {
      const newValues = getFormValues();
      const newNode = produce(node, (draft) => {
        draft.config = {
          ...newValues,
        };
        draft.validated = true;
      });
      await updateNode(newNode);
    } finally {
      setFormPending(false);
    }
  };

  return (
    <>
      <SharedNode.Block node={node} disabled={disabled} onClick={() => setDrawerOpen(true)}>
        {wrappedEl}
      </SharedNode.Block>

      <SharedNode.ConfigDrawer
        getConfigNewValues={getFormValues}
        node={node}
        open={drawerOpen}
        pending={formPending}
        onConfirm={handleDrawerConfirm}
        onOpenChange={(open) => setDrawerOpen(open)}
      >
        <NotifyNodeConfigForm ref={formRef} disabled={disabled} initialValues={node.config} />
      </SharedNode.ConfigDrawer>
    </>
  );
};

export default memo(NotifyNode);
