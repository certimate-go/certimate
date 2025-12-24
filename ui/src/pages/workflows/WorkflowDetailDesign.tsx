import { useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { FlowLayoutDefault } from "@flowgram.ai/fixed-layout-editor";
import { IconArrowBackUp, IconDots, IconTransferIn, IconTransferOut } from "@tabler/icons-react";
import { useDeepCompareEffect } from "ahooks";
import { Alert, App, Button, Card, Dropdown, Result, Space, theme } from "antd";
import { debounce } from "radash";

import Show from "@/components/Show";
import { WorkflowDesigner, type WorkflowDesignerInstance, WorkflowNodeDrawer, WorkflowToolbar } from "@/components/workflow/designer";
import WorkflowGraphExportModal from "@/components/workflow/WorkflowGraphExportModal";
import WorkflowGraphImportModal from "@/components/workflow/WorkflowGraphImportModal";
import { useAppSettings, useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";
import { unwrapErrMsg } from "@/utils/error";

const WorkflowDetailDesign = () => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();
  const { message, modal, notification } = App.useApp();

  const { appSettings: globalAppSettings } = useAppSettings();

  const { workflow, ...workflowStore } = useWorkflowStore(useZustandShallowSelector(["workflow", "orchestrate", "publish", "rollback"]));

  const workflowRollbackDisabled = useMemo(() => !workflow.hasDraft || !workflow.hasContent, [workflow.hasDraft, workflow.hasContent]);
  const workflowPublishDisabled = useMemo(() => !workflow.hasDraft, [workflow.hasDraft]);

  const designerRef = useRef<WorkflowDesignerInstance>(null);
  const designerPending = useRef(false); // 保存中时阻止刷新画布
  const [designerError, setDesignerError] = useState<unknown>();
  useDeepCompareEffect(() => {
    if (designerRef.current == null || designerRef.current.document.disposed) return;
    if (designerPending.current) return;

    try {
      const graph = workflow.graphDraft ?? { nodes: [] };
      designerRef.current!.document.fromJSON(graph);
      setDesignerError(void 0);
    } catch (err) {
      console.error(err);
      setDesignerError(err);
    }
  }, [workflow.graphDraft]);

  const { drawerProps: designerNodeDrawerProps, ...designerNodeDrawer } = WorkflowNodeDrawer.useDrawer();

  const handleDesignerDocumentChange = debounce({ delay: 300 }, async () => {
    if (designerRef.current == null || designerRef.current.document.disposed) return;

    designerPending.current = true;
    try {
      const graph = designerRef.current!.document.toJSON();
      await workflowStore.orchestrate(graph);
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    } finally {
      designerPending.current = false;
    }
  });

  const handleRollbackClick = () => {
    modal.confirm({
      title: t("workflow.detail.design.action.rollback.modal.title"),
      content: t("workflow.detail.design.action.rollback.modal.content"),
      onOk: async () => {
        try {
          await workflowStore.rollback();

          message.success(t("common.text.operation_succeeded"));
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  const handlePublishClick = async () => {
    if (!(await designerRef.current!.validateAllNodes())) {
      message.warning(t("workflow.detail.design.uncompleted_design.alert"));
      return;
    }

    modal.confirm({
      title: t("workflow.detail.design.action.publish.modal.title"),
      content: t("workflow.detail.design.action.publish.modal.content"),
      onOk: async () => {
        try {
          await workflowStore.publish();

          message.success(t("common.text.operation_succeeded"));
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  const { modalProps: graphImportModalProps, ...graphImportModal } = WorkflowGraphImportModal.useModal();
  const { modalProps: graphExportModalProps, ...graphExportModal } = WorkflowGraphExportModal.useModal();

  const handleImportClick = async () => {
    graphImportModal.open().then(async (graph) => {
      const loadingKey = Math.random().toString(36).substring(0, 8);
      message.loading({ key: loadingKey, content: t("common.text.saving"), duration: 0 });

      try {
        await workflowStore.orchestrate(graph);

        message.destroy(loadingKey);
        message.success(t("common.text.operation_succeeded"));
      } catch (err) {
        console.error(err);
        message.destroy(loadingKey);
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
      }
    });
  };

  const handleExportClick = () => {
    graphExportModal.open({ data: workflow.graphDraft! });
  };

  return (
    <div className="size-full">
      <Card
        className="size-full overflow-hidden"
        styles={{
          body: {
            position: "relative",
            height: "100%",
            padding: 0,
          },
        }}
      >
        <WorkflowDesigner
          ref={designerRef}
          defaultLayout={
            globalAppSettings.defaultWorkflowLayout === "horizontal"
              ? FlowLayoutDefault.HORIZONTAL_FIXED_LAYOUT
              : globalAppSettings.defaultWorkflowLayout === "vertical"
                ? FlowLayoutDefault.VERTICAL_FIXED_LAYOUT
                : void 0
          }
          onDocumentChange={handleDesignerDocumentChange}
          onNodeClick={(_, node) => designerNodeDrawer.open(node)}
        >
          <div className="absolute top-8 z-10 w-full px-4">
            <div className="container">
              <div className="flex items-center justify-end gap-4">
                <div className="flex flex-1 items-center justify-end gap-4 overflow-hidden">
                  <div className="flex-1 overflow-hidden">
                    <Show when={workflow.hasDraft!}>
                      <Alert showIcon title={<div className="truncate">{t("workflow.detail.design.unpublished_draft.alert")}</div>} type="warning" />
                    </Show>
                  </div>
                  <Space.Compact>
                    <Button disabled={workflowPublishDisabled} ghost type="primary" onClick={handlePublishClick}>
                      {t("workflow.detail.design.action.publish.button")}
                    </Button>
                    <Dropdown
                      menu={{
                        items: [
                          {
                            key: "rollback",
                            disabled: workflowRollbackDisabled,
                            label: t("workflow.detail.design.action.rollback.button"),
                            icon: <IconArrowBackUp size="1.25em" />,
                            onClick: handleRollbackClick,
                          },
                          {
                            type: "divider",
                          },
                          {
                            key: "import",
                            label: t("workflow.detail.design.action.import.button"),
                            icon: <IconTransferIn size="1.25em" />,
                            onClick: handleImportClick,
                          },
                          {
                            key: "export",
                            label: t("workflow.detail.design.action.export.button"),
                            icon: <IconTransferOut size="1.25em" />,
                            onClick: handleExportClick,
                          },
                        ],
                      }}
                      trigger={["click"]}
                    >
                      <Button icon={<IconDots size="1.25em" />} />
                    </Dropdown>
                  </Space.Compact>
                </div>
              </div>
            </div>
          </div>

          <div className="absolute bottom-8 z-10 w-full px-4">
            <div className="container">
              <div className="flex justify-end">
                <WorkflowToolbar
                  style={{
                    backgroundColor: themeToken.colorBgContainer,
                    borderRadius: themeToken.borderRadius,
                  }}
                />
              </div>
            </div>
          </div>

          {!!designerError && (
            <div className="absolute top-1/2 left-1/2 z-10 w-full -translate-1/2 px-4">
              <Result status="warning" title="Data corruption!" subTitle={`Error: ${unwrapErrMsg(designerError)}`} />
            </div>
          )}

          <WorkflowNodeDrawer {...designerNodeDrawerProps} />
        </WorkflowDesigner>

        <WorkflowGraphImportModal {...graphImportModalProps} />
        <WorkflowGraphExportModal {...graphExportModalProps} />
      </Card>
    </div>
  );
};

export default WorkflowDetailDesign;
