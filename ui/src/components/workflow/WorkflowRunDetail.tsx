import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { EditorState, FlowLayoutDefault } from "@flowgram.ai/fixed-layout-editor";
import { IconBrowserShare, IconBug, IconCheck, IconDots, IconDownload, IconSettings2, IconTransferOut } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { Alert, App, Button, Card, Divider, Dropdown, Empty, Skeleton, Table, type TableProps, Tooltip, Typography, theme } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import CertificateDetailDrawer from "@/components/certificate/CertificateDetailDrawer";
import Show from "@/components/Show";
import { type CertificateModel } from "@/domain/certificate";
import { WorkflowLogLevel, type WorkflowLogModel } from "@/domain/workflowLog";
import { WORKFLOW_RUN_STATUSES, type WorkflowRunModel } from "@/domain/workflowRun";
import { useBrowserTheme } from "@/hooks";
import { listByWorkflowRunId as listCertificatesByWorkflowRunId } from "@/repository/certificate";
import { listByWorkflowRunId as listLogsByWorkflowRunId } from "@/repository/workflowLog";
import { subscribe as subscribeWorkflowRun } from "@/repository/workflowRun";
import { mergeCls } from "@/utils/css";
import { unwrapErrMsg } from "@/utils/error";

import WorkflowDesigner from "./designer/Designer";
import WorkflowToolbar from "./designer/Toolbar";
import WorkflowGraphExportModal from "./WorkflowGraphExportModal";
import WorkflowStatus from "./WorkflowStatus";

export interface WorkflowRunDetailProps {
  className?: string;
  style?: React.CSSProperties;
  data: WorkflowRunModel;
}

const WorkflowRunDetail = ({ className, style, ...props }: WorkflowRunDetailProps) => {
  const { t } = useTranslation();

  const [innerData, setInnerData] = useState(props.data);
  const mergedData = useMemo(() => ({ ...props.data, ...innerData }), [innerData, props.data]);

  const unsubscriberRef = useRef<() => void>();
  useEffect(() => {
    if (props.data.status === WORKFLOW_RUN_STATUSES.PENDING || props.data.status === WORKFLOW_RUN_STATUSES.PROCESSING) {
      subscribeWorkflowRun(props.data.id, (cb) => {
        setInnerData(cb.record);

        if (cb.record.status !== WORKFLOW_RUN_STATUSES.PENDING && cb.record.status !== WORKFLOW_RUN_STATUSES.PROCESSING) {
          unsubscriberRef.current?.();
          unsubscriberRef.current = undefined;
        }
      }).then((unsubscriber) => {
        unsubscriberRef.current = unsubscriber;
      });
    }

    return () => {
      unsubscriberRef.current?.();
      unsubscriberRef.current = undefined;
    };
  }, [props.data.id, props.data.status]);

  return (
    <div className={className} style={style}>
      <Alert
        showIcon
        title={
          <div className="text-xs">
            {mergedData.endedAt
              ? t("workflow_run.base.description_with_time_cost", {
                  trigger: t(`workflow_run.base.trigger.${mergedData.trigger}`),
                  startedAt: dayjs(mergedData.startedAt).format("YYYY-MM-DD HH:mm:ss"),
                  timeCost: dayjs(mergedData.endedAt).diff(dayjs(mergedData.startedAt), "second") + "s",
                })
              : t("workflow_run.base.description", {
                  trigger: t(`workflow_run.base.trigger.${mergedData.trigger}`),
                  startedAt: dayjs(mergedData.startedAt).format("YYYY-MM-DD HH:mm:ss"),
                })}
          </div>
        }
        type={
          {
            [WORKFLOW_RUN_STATUSES.SUCCEEDED]: "success" as const,
            [WORKFLOW_RUN_STATUSES.FAILED]: "error" as const,
            [WORKFLOW_RUN_STATUSES.CANCELED]: "warning" as const,
          }[mergedData.status] ?? ("info" as const)
        }
      />
      {!!mergedData.error && (
        <Alert
          className="mt-1"
          icon={<IconBug size="1em" color="var(--color-error)" />}
          showIcon
          title={<div className="text-xs text-error">{mergedData.error}</div>}
        />
      )}

      <div className="mt-8">
        <Typography.Title level={5}>{t("workflow_run.process")}</Typography.Title>
        <WorkflowRunProcess runData={mergedData} />
      </div>

      <div className="mt-8">
        <Typography.Title level={5}>{t("workflow_run.logs")}</Typography.Title>
        <WorkflowRunLogs runId={mergedData.id} runStatus={mergedData.status} />
      </div>

      <Show when={mergedData.status === WORKFLOW_RUN_STATUSES.SUCCEEDED}>
        <div className="mt-8">
          <Typography.Title level={5}>{t("workflow_run.artifacts")}</Typography.Title>
          <WorkflowRunArtifacts runId={mergedData.id} />
        </div>
      </Show>
    </div>
  );
};

const WorkflowRunProcess = ({ runData }: { runData: WorkflowRunModel }) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { modalProps: graphExportModalProps, ...graphExportModal } = WorkflowGraphExportModal.useModal();

  const handleExportClick = () => {
    graphExportModal.open({ data: runData.graph! });
  };

  return (
    <>
      <Card
        className="size-full overflow-hidden"
        styles={{
          body: {
            position: "relative",
            height: "240px",
            padding: 0,
            cursor: "grab",
          },
        }}
      >
        <WorkflowDesigner
          defaultEditorState={EditorState.STATE_MOUSE_FRIENDLY_SELECT.id}
          defaultLayout={FlowLayoutDefault.HORIZONTAL_FIXED_LAYOUT}
          initialData={runData.graph}
          readonly
        >
          <div className="absolute bottom-4 z-10 w-full px-4">
            <div className="container">
              <div className="flex items-center justify-end gap-2">
                <WorkflowToolbar
                  style={{
                    backgroundColor: themeToken.colorBgContainer,
                    borderRadius: themeToken.borderRadius,
                  }}
                  size="small"
                  showMouseState={false}
                  showLayout={false}
                  showMinimap={false}
                  showZoomLevel={false}
                />

                <Dropdown
                  menu={{
                    items: [
                      {
                        key: "export",
                        label: t("workflow_run.process.menu.export"),
                        icon: <IconTransferOut size="1.25em" />,
                        onClick: handleExportClick,
                      },
                    ],
                  }}
                  trigger={["click"]}
                >
                  <Button icon={<IconDots size="1.25em" />} size="small" />
                </Dropdown>
              </div>
            </div>
          </div>
        </WorkflowDesigner>
      </Card>

      <WorkflowGraphExportModal {...graphExportModalProps} />
    </>
  );
};

const WorkflowRunLogs = ({ runId, runStatus }: { runId: string; runStatus: string }) => {
  const { t } = useTranslation();

  const { theme: browserTheme } = useBrowserTheme();

  type Log = Pick<WorkflowLogModel, "timestamp" | "level" | "message" | "data">;
  type LogGroup = { id: string; name: string; records: Log[] };
  const [listData, setListData] = useState<LogGroup[]>([]);
  const { loading, ...req } = useRequest(
    () => {
      return listLogsByWorkflowRunId(runId);
    },
    {
      refreshDeps: [runId, runStatus],
      pollingInterval: 1500,
      pollingWhenHidden: false,
      throttleWait: 500,
      onSuccess: (res) => {
        if (res.items.length === listData.flatMap((e) => e.records).length) return;

        setListData(
          res.items.reduce((acc, e) => {
            let group = acc.at(-1);
            if (!group || group.id !== e.nodeId) {
              group = { id: e.nodeId, name: e.nodeName, records: [] };
              acc.push(group);
            }
            group.records.push({ timestamp: e.timestamp, level: e.level, message: e.message, data: e.data });
            return acc;
          }, [] as LogGroup[])
        );
      },
      onFinally: () => {
        if (runStatus === WORKFLOW_RUN_STATUSES.PENDING || runStatus === WORKFLOW_RUN_STATUSES.PROCESSING) {
          req.cancel();
        }
      },
      onError: (err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);

        throw err;
      },
    }
  );

  const [showTimestamp, setShowTimestamp] = useState(true);
  const [showWhitespace, setShowWhitespace] = useState(true);

  const renderLogRecord = (record: Log) => {
    let timestamp = dayjs(record.timestamp).format("YYYY-MM-DD HH:mm:ss");
    timestamp = `[${timestamp}]`;

    let message = <>{record.message}</>;
    if (record.data != null && Object.keys(record.data).length > 0) {
      message = (
        <details>
          <summary>{record.message}</summary>
          {Object.entries(record.data).map(([key, value]) => (
            <div key={key} className="flex space-x-2" style={{ wordBreak: "break-word" }}>
              <div className="whitespace-nowrap">{key}:</div>
              <div className={showWhitespace ? "whitespace-normal" : "whitespace-pre-line"}>{JSON.stringify(value)}</div>
            </div>
          ))}
        </details>
      );
    }

    return (
      <div className="flex space-x-2" style={{ wordBreak: "break-word" }}>
        {showTimestamp && <div className="font-mono whitespace-nowrap text-stone-400">{timestamp}</div>}
        <div
          className={mergeCls(
            "flex-1 font-mono",
            { ["whitespace-pre-line"]: !showWhitespace },
            record.level < WorkflowLogLevel.Info
              ? "text-stone-400"
              : record.level < WorkflowLogLevel.Warn
                ? ""
                : record.level < WorkflowLogLevel.Error
                  ? "text-warning"
                  : "text-error"
          )}
        >
          {message}
        </div>
      </div>
    );
  };

  const handleDownloadClick = () => {
    const NEWLINE = "\n";
    const logstr = listData
      .map((group) => {
        const escape = (str: string) => str.replaceAll("\r", "\\r").replaceAll("\n", "\\n");
        return (
          `#${group.id} ${group.name}` +
          NEWLINE +
          group.records
            .map((record) => {
              const datetime = dayjs(record.timestamp).format("YYYY-MM-DDTHH:mm:ss.SSSZ");
              const level =
                record.level < WorkflowLogLevel.Info
                  ? "DBUG"
                  : record.level < WorkflowLogLevel.Warn
                    ? "INFO"
                    : record.level < WorkflowLogLevel.Error
                      ? "WARN"
                      : "ERRO";
              const message = record.message;
              const data = record.data && Object.keys(record.data).length > 0 ? JSON.stringify(record.data) : "";
              return `[${datetime}] [${level}] ${escape(message)} ${escape(data)}`.trim();
            })
            .join(NEWLINE)
        );
      })
      .join(NEWLINE + NEWLINE);
    const blob = new Blob([logstr], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `certimate_workflow_run_#${runId}_logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
    a.remove();
  };

  return (
    <div className="rounded-md bg-black text-stone-200">
      <div className="flex items-center gap-2 p-4">
        <div className="grow overflow-hidden">
          <WorkflowStatus value={runStatus} />
        </div>
        <div>
          <Dropdown
            menu={{
              items: [
                {
                  key: "show-timestamp",
                  label: t("workflow_run.logs.menu.show_timestamps"),
                  icon: <IconCheck className={showTimestamp ? "visible" : "invisible"} size="1.25em" />,
                  onClick: () => setShowTimestamp(!showTimestamp),
                },
                {
                  key: "show-whitespace",
                  label: t("workflow_run.logs.menu.show_whitespaces"),
                  icon: <IconCheck className={showWhitespace ? "visible" : "invisible"} size="1.25em" />,
                  onClick: () => setShowWhitespace(!showWhitespace),
                },
                {
                  type: "divider",
                },
                {
                  key: "download-logs",
                  label: t("workflow_run.logs.menu.download_logs"),
                  icon: <IconDownload className="invisible" size="1.25em" />,
                  onClick: handleDownloadClick,
                },
              ],
            }}
            trigger={["click"]}
          >
            <Button color="primary" icon={<IconSettings2 size="1.25em" />} ghost={browserTheme === "light"} />
          </Dropdown>
        </div>
      </div>

      <Divider className="my-0 bg-stone-800" />

      <div className="min-h-8 px-4 py-2">
        <Show when={!loading || listData.length > 0} fallback={<Skeleton />}>
          {listData.map((group) => {
            return (
              <div className="mb-3">
                <div className="truncate text-xs leading-loose">
                  <span className="font-mono text-stone-400">{`#${group.id}\u00A0`}</span>
                  <span>{group.name}</span>
                </div>
                <div className="flex flex-col text-xs leading-relaxed">{group.records.map((record) => renderLogRecord(record))}</div>
              </div>
            );
          })}
        </Show>
      </div>
    </div>
  );
};

const WorkflowRunArtifacts = ({ runId }: { runId: string }) => {
  const { t } = useTranslation();

  const { notification } = App.useApp();

  const tableColumns: TableProps<CertificateModel>["columns"] = [
    {
      key: "$index",
      align: "center",
      fixed: "left",
      width: 50,
      render: (_, __, index) => index + 1,
    },
    {
      key: "type",
      title: t("workflow_run_artifact.props.type"),
      render: () => t("workflow_run_artifact.props.type.certificate"),
    },
    {
      key: "name",
      title: t("workflow_run_artifact.props.name"),
      render: (_, record) => {
        return (
          <div className="max-w-full truncate">
            <Typography.Text delete={!!record.deleted} ellipsis>
              {record.subjectAltNames}
            </Typography.Text>
          </div>
        );
      },
    },
    {
      key: "$action",
      align: "end",
      width: 32,
      render: (_, record) => (
        <div className="flex items-center justify-end">
          <CertificateDetailDrawer
            data={record}
            trigger={
              <Tooltip title={t("common.button.view")}>
                <Button color="primary" disabled={!!record.deleted} icon={<IconBrowserShare size="1.25em" />} variant="text" />
              </Tooltip>
            }
          />
        </div>
      ),
    },
  ];
  const [tableData, setTableData] = useState<CertificateModel[]>([]);
  const { loading } = useRequest(
    () => {
      return listCertificatesByWorkflowRunId(runId);
    },
    {
      refreshDeps: [runId],
      onSuccess: (res) => {
        setTableData(res.items);
      },
      onError: (err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

        throw err;
      },
    }
  );

  return (
    <Table<CertificateModel>
      columns={tableColumns}
      dataSource={tableData}
      loading={loading}
      locale={{
        emptyText: <Empty description={t("common.text.nodata")} image={Empty.PRESENTED_IMAGE_SIMPLE} />,
      }}
      pagination={false}
      rowKey={(record) => record.id}
      size="small"
    />
  );
};

export default WorkflowRunDetail;
