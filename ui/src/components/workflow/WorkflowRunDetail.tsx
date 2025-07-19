import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  CheckCircleOutlined as CheckCircleOutlinedIcon,
  ClockCircleOutlined as ClockCircleOutlinedIcon,
  CloseCircleOutlined as CloseCircleOutlinedIcon,
  StopOutlined as StopOutlinedIcon,
  SyncOutlined as SyncOutlinedIcon,
} from "@ant-design/icons";
import { IconBrowserShare, IconCheck, IconChevronRight, IconDownload, IconSettings2 } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import {
  Button,
  Collapse,
  Divider,
  Dropdown,
  Empty,
  Flex,
  Skeleton,
  Space,
  Spin,
  Table,
  type TableProps,
  Tooltip,
  Typography,
  notification,
  theme,
} from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import CertificateDetailDrawer from "@/components/certificate/CertificateDetailDrawer";
import Show from "@/components/Show";
import { type CertificateModel } from "@/domain/certificate";
import { type WorkflowLogModel } from "@/domain/workflowLog";
import { WORKFLOW_RUN_STATUSES, type WorkflowRunModel } from "@/domain/workflowRun";
import { useBrowserTheme } from "@/hooks";
import { listByWorkflowRunId as listCertificatesByWorkflowRunId } from "@/repository/certificate";
import { listByWorkflowRunId as listLogsByWorkflowRunId } from "@/repository/workflowLog";
import { mergeCls } from "@/utils/css";
import { getErrMsg } from "@/utils/error";

export type WorkflowRunDetailProps = {
  className?: string;
  style?: React.CSSProperties;
  data: WorkflowRunModel;
};

const WorkflowRunDetail = ({ data, ...props }: WorkflowRunDetailProps) => {
  return (
    <div {...props}>
      <Show when={!!data}>
        <WorkflowRunLogs runId={data.id} runStatus={data.status} />
      </Show>

      <Show when={!!data && data.status === WORKFLOW_RUN_STATUSES.SUCCEEDED}>
        <Divider />
        <WorkflowRunArtifacts runId={data.id} />
      </Show>
    </div>
  );
};

const WorkflowRunLogs = ({ runId, runStatus }: { runId: string; runStatus: string }) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();
  const { theme: browserTheme } = useBrowserTheme();

  type Log = Pick<WorkflowLogModel, "timestamp" | "level" | "message" | "data">;
  type LogGroup = { id: string; name: string; records: Log[] };
  const [listData, setListData] = useState<LogGroup[]>([]);
  const { loading } = useRequest(
    () => {
      return listLogsByWorkflowRunId(runId);
    },
    {
      refreshDeps: [runId, runStatus],
      pollingInterval: runStatus === WORKFLOW_RUN_STATUSES.PENDING || runStatus === WORKFLOW_RUN_STATUSES.RUNNING ? 3000 : 0,
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

  const renderBadge = () => {
    switch (runStatus) {
      case WORKFLOW_RUN_STATUSES.PENDING:
        return (
          <Flex gap="small">
            <ClockCircleOutlinedIcon />
            {t("workflow_run.props.status.pending")}
          </Flex>
        );
      case WORKFLOW_RUN_STATUSES.RUNNING:
        return (
          <Flex gap="small" style={{ color: themeToken.colorInfo }}>
            <SyncOutlinedIcon spin />
            {t("workflow_run.props.status.running")}
          </Flex>
        );
      case WORKFLOW_RUN_STATUSES.SUCCEEDED:
        return (
          <Flex gap="small" style={{ color: themeToken.colorSuccess }}>
            <CheckCircleOutlinedIcon />
            {t("workflow_run.props.status.succeeded")}
          </Flex>
        );
      case WORKFLOW_RUN_STATUSES.FAILED:
        return (
          <Flex gap="small" style={{ color: themeToken.colorError }}>
            <CloseCircleOutlinedIcon />
            {t("workflow_run.props.status.failed")}
          </Flex>
        );
      case WORKFLOW_RUN_STATUSES.CANCELED:
        return (
          <Flex gap="small" style={{ color: themeToken.colorWarning }}>
            <StopOutlinedIcon />
            {t("workflow_run.props.status.canceled")}
          </Flex>
        );
    }

    return <></>;
  };

  const renderRecord = (record: Log) => {
    let message = <>{record.message}</>;
    if (record.data != null && Object.keys(record.data).length > 0) {
      message = (
        <details>
          <summary>{record.message}</summary>
          {Object.entries(record.data).map(([key, value]) => (
            <div key={key} className="flex space-x-2" style={{ wordBreak: "break-word" }}>
              <div className="whitespace-nowrap">{key}:</div>
              <div className={!showWhitespace ? "whitespace-pre-line" : ""}>{JSON.stringify(value)}</div>
            </div>
          ))}
        </details>
      );
    }

    return (
      <div className="flex space-x-2 text-xs" style={{ wordBreak: "break-word" }}>
        {showTimestamp ? <div className="whitespace-nowrap text-stone-400">[{dayjs(record.timestamp).format("YYYY-MM-DD HH:mm:ss")}]</div> : <></>}
        <div
          className={mergeCls(
            "font-mono",
            record.level === "DEBUG" ? "text-stone-400" : "",
            record.level === "WARN" ? "text-yellow-600" : "",
            record.level === "ERROR" ? "text-red-600" : "",
            !showWhitespace ? "whitespace-pre-line" : ""
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
          group.name +
          NEWLINE +
          group.records
            .map((record) => {
              const datetime = dayjs(record.timestamp).format("YYYY-MM-DDTHH:mm:ss.SSSZ");
              const level = record.level;
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
    <>
      <Typography.Title level={5}>{t("workflow_run.logs")}</Typography.Title>
      <div className="rounded-md bg-black text-stone-200">
        <div className="flex items-center gap-2 p-4">
          <div className="grow overflow-hidden">{renderBadge()}</div>
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

        <Show
          when={!loading || listData.length > 0}
          fallback={
            <Spin spinning>
              <Skeleton />
            </Spin>
          }
        >
          <div className="py-2">
            <Collapse
              style={{ color: "inherit" }}
              bordered={false}
              defaultActiveKey={listData.map((group) => group.id)}
              expandIcon={({ isActive }) => <IconChevronRight className={mergeCls(isActive ? "" : "rotate-90", "transition-transform")} size="1.25em" />}
              items={listData.map((group) => {
                return {
                  key: group.id,
                  classNames: {
                    header: "text-sm text-stone-200",
                    body: "text-stone-200",
                  },
                  style: { color: "inherit", border: "none" },
                  styles: {
                    header: { color: "inherit" },
                  },
                  label: group.name,
                  children: <div className="flex flex-col space-y-1">{group.records.map((record) => renderRecord(record))}</div>,
                };
              })}
            />
          </div>
        </Show>
      </div>
    </>
  );
};

const WorkflowRunArtifacts = ({ runId }: { runId: string }) => {
  const { t } = useTranslation();

  const [notificationApi, NotificationContextHolder] = notification.useNotification();

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
      ellipsis: true,
      render: (_, record) => {
        return (
          <Typography.Text delete={!!record.deleted} ellipsis>
            {record.subjectAltNames}
          </Typography.Text>
        );
      },
    },
    {
      key: "$action",
      align: "end",
      width: 120,
      render: (_, record) => (
        <Space.Compact>
          <CertificateDetailDrawer
            data={record}
            trigger={
              <Tooltip title={t("certificate.action.view")}>
                <Button color="primary" disabled={!!record.deleted} icon={<IconBrowserShare size="1.25em" />} variant="text" />
              </Tooltip>
            }
          />
        </Space.Compact>
      ),
    },
  ];
  const [tableData, setTableData] = useState<CertificateModel[]>([]);
  const { loading: tableLoading } = useRequest(
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
        notificationApi.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      },
    }
  );

  return (
    <>
      {NotificationContextHolder}

      <Typography.Title level={5}>{t("workflow_run.artifacts")}</Typography.Title>
      <Table<CertificateModel>
        columns={tableColumns}
        dataSource={tableData}
        loading={tableLoading}
        locale={{
          emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />,
        }}
        pagination={false}
        rowKey={(record) => record.id}
        size="small"
      />
    </>
  );
};

export default WorkflowRunDetail;
