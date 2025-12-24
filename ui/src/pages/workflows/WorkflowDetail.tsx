import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Outlet, useLocation, useNavigate, useParams } from "react-router-dom";
import { IconEdit, IconHistory, IconPlayerPlay, IconRobot } from "@tabler/icons-react";
import { useSize } from "ahooks";
import { App, Button, Input, type InputRef, Segmented, Skeleton } from "antd";

import { startRun as startWorkflowRun } from "@/api/workflows";
import Show from "@/components/Show";
import { WORKFLOW_RUN_STATUSES } from "@/domain/workflowRun";
import { useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";
import { mergeCls } from "@/utils/css";
import { unwrapErrMsg } from "@/utils/error";

const WorkflowDetail = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { message, modal, notification } = App.useApp();

  const { id: workflowId } = useParams();
  const { workflow, initialized, ...workflowState } = useWorkflowStore(useZustandShallowSelector(["workflow", "initialized", "init", "destroy", "setEnabled"]));
  useEffect(() => {
    Promise.try(() => workflowState.init(workflowId!)).catch((err) => {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    });

    return () => {
      workflowState.destroy();
    };
  }, [workflowId]);

  const divHeaderRef = useRef<HTMLDivElement>(null);
  const divHeaderSize = useSize(divHeaderRef);

  const tabs = [
    ["design", "workflow.detail.design.tab", <IconRobot size="1em" />],
    ["runs", "workflow.detail.runs.tab", <IconHistory size="1em" />],
  ] satisfies [string, string, React.ReactElement][];
  const [tabValue, setTabValue] = useState<string>(() => location.pathname.split("/")[3]);
  useEffect(() => {
    const subpath = location.pathname.split("/")[3];
    if (!subpath) {
      navigate(`/workflows/${workflowId}/${tabs[0][0]}`, { replace: true });
      return;
    }

    setTabValue(subpath);
  }, [location.pathname, workflowId]);

  const handleTabChange = (value: string) => {
    setTabValue(value);
    navigate(`/workflows/${workflowId}/${value}`);
  };

  const runButtonDisabled = useMemo(() => !workflow.hasContent, [workflow]);
  const [runButtonLoading, setRunButtonLoading] = useState(false);
  useEffect(() => {
    const running = workflow.lastRunStatus === WORKFLOW_RUN_STATUSES.PENDING || workflow.lastRunStatus === WORKFLOW_RUN_STATUSES.PROCESSING;
    setRunButtonLoading(running);
  }, [workflow.lastRunStatus]);

  const handleRunClick = () => {
    const { promise, resolve } = Promise.withResolvers();
    if (workflow.hasDraft) {
      modal.confirm({
        title: t("workflow.action.execute.modal.title"),
        content: t("workflow.action.execute.modal.content"),
        onOk: () => resolve(void 0),
      });
    } else {
      resolve(void 0);
    }

    promise.then(async () => {
      try {
        setRunButtonLoading(true);

        await startWorkflowRun(workflow.id);

        message.info(t("workflow.action.execute.prompt"));
      } catch (err) {
        setRunButtonLoading(false);

        console.error(err);
        message.warning(t("common.text.operation_failed"));
      }
    });
  };

  const handleActiveClick = async () => {
    try {
      if (!workflow.enabled && !workflow.graphContent) {
        message.warning(t("workflow.action.enable.errmsg.unpublished"));
        return;
      }

      await workflowState.setEnabled(!workflow.enabled);
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    }
  };

  return (
    <div className="flex size-full flex-col">
      <div className="px-6 py-4" ref={divHeaderRef}>
        <div className="relative z-11 container flex justify-between gap-4">
          <div className="flex-1">
            <WorkflowDetailBaseName />
            <WorkflowDetailBaseDescription />

            <div className="absolute -bottom-12 left-1/2 z-1 -translate-x-1/2">
              <Segmented
                className="shadow"
                options={tabs.map(([key, label, icon]) => ({
                  value: key,
                  label: <span className="px-2 text-sm">{t(label)}</span>,
                  icon: (
                    <span className="anticon scale-125" role="img">
                      {icon}
                    </span>
                  ),
                }))}
                size="large"
                value={tabValue}
                onChange={handleTabChange}
              />
            </div>
          </div>
          <div className="py-2">
            <Show when={initialized}>
              <div className="flex items-center gap-2">
                <Button onClick={handleActiveClick}>{workflow.enabled ? t("workflow.action.disable.button") : t("workflow.action.enable.button")}</Button>
                <Button disabled={runButtonDisabled} icon={<IconPlayerPlay size="1.25em" />} loading={runButtonLoading} type="primary" onClick={handleRunClick}>
                  {t("workflow.action.execute.button")}
                </Button>
              </div>
            </Show>
          </div>
        </div>
      </div>

      <div
        className="flex-1 p-4"
        style={{
          minHeight: `calc(max(360px, 100% - ${divHeaderSize?.height ?? 0}px))`,
        }}
      >
        <Show
          when={initialized}
          fallback={
            <div className="container pt-12">
              <Skeleton active />
            </div>
          }
        >
          <Outlet />
        </Show>
      </div>
    </div>
  );
};

const WorkflowDetailBaseName = () => {
  const { t } = useTranslation();

  const { notification } = App.useApp();

  const { workflow, initialized, ...workflowStore } = useWorkflowStore(useZustandShallowSelector(["workflow", "initialized", "setName"]));

  const inputRef = useRef<InputRef>(null);
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState("");

  useEffect(() => {
    setEditing(false);
  }, [workflow.id]);

  const handleEditClick = () => {
    setEditing(true);
    setValue(workflow.name);
    setTimeout(() => {
      inputRef.current?.focus({ cursor: "all" });
    }, 0);
  };

  const handleValueChange = (value: string) => {
    setValue(value);
  };

  const handleValueConfirm = async (value: string) => {
    value = value.trim();
    if (!value || value === (workflow.name || "")) {
      setEditing(false);
      return;
    }

    setEditing(false);

    try {
      await workflowStore.setName(value);
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

      throw err;
    }
  };

  return (
    <div className="group/input relative flex items-center gap-1">
      <h1 className={mergeCls("break-all", { invisible: editing })}>
        <Show when={initialized} fallback={"\u00A0"}>
          {workflow.name || t("workflow.detail.baseinfo.name.placeholder")}
        </Show>
      </h1>
      <Show when={initialized}>
        <Button
          className={mergeCls("mb-2 opacity-0 transition-opacity group-hover/input:opacity-100", {
            invisible: editing,
          })}
          icon={<IconEdit size="1.25em" stroke="1.25" />}
          type="text"
          onClick={handleEditClick}
        />
      </Show>
      <Input
        className={mergeCls("absolute top-0 left-0", editing ? "block" : "hidden")}
        ref={inputRef}
        maxLength={100}
        placeholder={t("workflow.detail.baseinfo.name.placeholder")}
        size="large"
        value={value}
        variant="filled"
        onBlur={(e) => handleValueConfirm(e.target.value)}
        onChange={(e) => handleValueChange(e.target.value)}
        onPressEnter={(e) => e.currentTarget.blur()}
      />
    </div>
  );
};

const WorkflowDetailBaseDescription = () => {
  const { t } = useTranslation();

  const { notification } = App.useApp();

  const { workflow, initialized, ...workflowStore } = useWorkflowStore(useZustandShallowSelector(["workflow", "initialized", "setDescription"]));

  const inputRef = useRef<InputRef>(null);
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState("");

  useEffect(() => {
    setEditing(false);
  }, [workflow.id]);

  const handleEditClick = () => {
    setEditing(true);
    setValue(workflow.description || "");
    setTimeout(() => {
      inputRef.current?.focus({ cursor: "all" });
    }, 0);
  };

  const handleValueChange = (value: string) => {
    setValue(value);
  };

  const handleValueConfirm = async (value: string) => {
    value = value.trim();
    if (!value || value === (workflow.description || "")) {
      setEditing(false);
      return;
    }

    setEditing(false);

    try {
      await workflowStore.setDescription(value);
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

      throw err;
    }
  };

  return (
    <div className="group/input relative flex items-center gap-1">
      <p className={mergeCls("text-base text-gray-500", { invisible: editing })}>
        <Show when={initialized} fallback={"\u00A0"}>
          {workflow.description || t("workflow.detail.baseinfo.description.placeholder")}
        </Show>
      </p>
      <Show when={initialized}>
        <Button
          className={mergeCls("mb-4 opacity-0 transition-opacity group-hover/input:opacity-100", {
            invisible: editing,
          })}
          icon={<IconEdit size="1.25em" stroke="1.25" />}
          type="text"
          onClick={handleEditClick}
        />
      </Show>
      <Input
        className={mergeCls("absolute top-0 left-0", editing ? "block" : "hidden")}
        ref={inputRef}
        maxLength={100}
        placeholder={t("workflow.detail.baseinfo.description.placeholder")}
        value={value}
        variant="filled"
        onBlur={(e) => handleValueConfirm(e.target.value)}
        onChange={(e) => handleValueChange(e.target.value)}
        onPressEnter={(e) => e.currentTarget.blur()}
      />
    </div>
  );
};

export default WorkflowDetail;
