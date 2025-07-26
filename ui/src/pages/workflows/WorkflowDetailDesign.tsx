import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconArrowBackUp, IconDots, IconPlayerPlay } from "@tabler/icons-react";
import { Alert, App, Button, Card, Dropdown, Space } from "antd";
import { isEqual } from "radash";

import { startRun as startWorkflowRun } from "@/api/workflows";
import Show from "@/components/Show";
import WorkflowElementsContainer from "@/components/workflow/WorkflowElementsContainer";
import { isAllNodesValidated } from "@/domain/workflow";
import { WORKFLOW_RUN_STATUSES } from "@/domain/workflowRun";
import { useZustandShallowSelector } from "@/hooks";
import { subscribe as subscribeWorkflow } from "@/repository/workflow";
import { useWorkflowStore } from "@/stores/workflow";
import { getErrMsg } from "@/utils/error";

const WorkflowDetailDesign = () => {
  const { t } = useTranslation();

  const { message, modal, notification } = App.useApp();

  const { workflow, ...workflowState } = useWorkflowStore(useZustandShallowSelector(["workflow", "init", "publish", "rollback"]));

  const [isPendingOrRunning, setIsPendingOrRunning] = useState(false);
  const [allowRollback, setAllowRollback] = useState(false);
  const [allowPublish, setAllowPublish] = useState(false);
  const [allowRun, setAllowRun] = useState(false);

  useEffect(() => {
    const pending = workflow.lastRunStatus === WORKFLOW_RUN_STATUSES.PENDING || workflow.lastRunStatus === WORKFLOW_RUN_STATUSES.RUNNING;
    setIsPendingOrRunning(pending);
  }, [workflow]);

  useEffect(() => {
    if (isPendingOrRunning) {
      let unsubscribeFn: Awaited<ReturnType<typeof subscribeWorkflow>> | undefined = undefined;
      subscribeWorkflow(workflow.id, (cb) => {
        if (cb.record.lastRunStatus !== WORKFLOW_RUN_STATUSES.PENDING && cb.record.lastRunStatus !== WORKFLOW_RUN_STATUSES.RUNNING) {
          setIsPendingOrRunning(false);
          unsubscribeFn?.();
        }
      }).then((res) => {
        unsubscribeFn = res;
      });

      return () => {
        unsubscribeFn?.();
      };
    }
  }, [workflow.id, isPendingOrRunning]);

  useEffect(() => {
    const hasContent = !!workflow.content;
    const hasChanges = workflow.hasDraft! || !isEqual(workflow.draft, workflow.content);
    setAllowRollback(!isPendingOrRunning && hasContent && hasChanges);
    setAllowPublish(!isPendingOrRunning && hasChanges);
    setAllowRun(hasContent);
  }, [workflow.content, workflow.draft, workflow.hasDraft, isPendingOrRunning]);

  const handleRollbackClick = () => {
    modal.confirm({
      title: t("workflow.action.rollback.modal.title"),
      content: t("workflow.action.rollback.modal.content"),
      onOk: async () => {
        try {
          await workflowState.rollback();

          message.success(t("common.text.operation_succeeded"));
        } catch (err) {
          console.error(err);
          notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handlePublishClick = () => {
    if (!isAllNodesValidated(workflow.draft!)) {
      message.warning(t("workflow.action.publish.errmsg.uncompleted"));
      return;
    }

    modal.confirm({
      title: t("workflow.action.publish.modal.title"),
      content: t("workflow.action.publish.modal.content"),
      onOk: async () => {
        try {
          await workflowState.publish();

          message.success(t("common.text.operation_succeeded"));
        } catch (err) {
          console.error(err);
          notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleRunClick = () => {
    const { promise, resolve, reject } = Promise.withResolvers();
    if (workflow.hasDraft) {
      modal.confirm({
        title: t("workflow.action.run.modal.title"),
        content: t("workflow.action.run.modal.content"),
        onOk: () => resolve(void 0),
        onCancel: () => reject(),
      });
    } else {
      resolve(void 0);
    }

    promise.then(async () => {
      let unsubscribeFn: Awaited<ReturnType<typeof subscribeWorkflow>> | undefined = undefined;

      try {
        setIsPendingOrRunning(true);

        // subscribe before running workflow
        unsubscribeFn = await subscribeWorkflow(workflow.id, (e) => {
          if (e.record.lastRunStatus !== WORKFLOW_RUN_STATUSES.PENDING && e.record.lastRunStatus !== WORKFLOW_RUN_STATUSES.RUNNING) {
            setIsPendingOrRunning(false);
            unsubscribeFn?.();
          }
        });

        await startWorkflowRun(workflow.id);

        message.info(t("workflow.action.run.prompt"));
      } catch (err) {
        setIsPendingOrRunning(false);
        unsubscribeFn?.();

        console.error(err);
        message.warning(t("common.text.operation_failed"));
      }
    });
  };

  return (
    <div className="min-h-[360px] flex-1 overflow-hidden">
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
        <div className="pt-9">
          <div className="absolute inset-x-6 z-2 mx-auto flex max-w-320 items-center justify-between gap-4">
            <div className="flex-1 overflow-hidden">
              <Show when={workflow.hasDraft!}>
                <Alert message={<div className="truncate">{t("workflow.detail.design.draft.alert")}</div>} showIcon type="warning" />
              </Show>
            </div>
            <div className="flex justify-end">
              <Space>
                <Button disabled={!allowRun} icon={<IconPlayerPlay size="1.25em" />} loading={isPendingOrRunning} type="primary" onClick={handleRunClick}>
                  {t("workflow.action.run.button")}
                </Button>
                <Space.Compact>
                  <Button color="primary" disabled={!allowPublish} variant="outlined" onClick={handlePublishClick}>
                    {t("workflow.action.publish.button")}
                  </Button>
                  <Dropdown
                    menu={{
                      items: [
                        {
                          key: "rollback",
                          disabled: !allowRollback,
                          label: t("workflow.action.rollback.button"),
                          icon: <IconArrowBackUp size="1.25em" />,
                          onClick: handleRollbackClick,
                        },
                      ],
                    }}
                    trigger={["click"]}
                  >
                    <Button color="primary" icon={<IconDots size="1.25em" />} variant="outlined" />
                  </Dropdown>
                </Space.Compact>
              </Space>
            </div>
          </div>

          <WorkflowElementsContainer className="pt-12" />
        </div>
      </Card>
    </div>
  );
};

export default WorkflowDetailDesign;
