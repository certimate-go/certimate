import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconBrowserShare, IconHistory, IconPlayerPause, IconTrash } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { Alert, App, Button, Skeleton, Table, type TableProps, Tooltip } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import { cancelRun as cancelWorkflowRun } from "@/api/workflows";
import Empty from "@/components/Empty";
import WorkflowRunDetailDrawer from "@/components/workflow/WorkflowRunDetailDrawer";
import WorkflowStatusTag from "@/components/workflow/WorkflowStatusTag";
import { WORKFLOW_TRIGGERS } from "@/domain/workflow";
import { WORKFLOW_RUN_STATUSES, type WorkflowRunModel } from "@/domain/workflowRun";
import { useAppSettings, useZustandShallowSelector } from "@/hooks";
import {
  list as listWorkflowRuns,
  remove as removeWorkflowRun,
  subscribe as subscribeWorkflowRun,
  unsubscribe as unsubscribeWorkflowRun,
} from "@/repository/workflowRun";
import { useWorkflowStore } from "@/stores/workflow";
import { getErrMsg } from "@/utils/error";

const WorkflowDetailRuns = () => {
  const { t } = useTranslation();

  const { modal, notification } = App.useApp();

  const { appSettings: globalAppSettings } = useAppSettings();

  const { workflow } = useWorkflowStore(useZustandShallowSelector(["workflow"]));

  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(globalAppSettings.defaultPerPage!);

  const [tableData, setTableData] = useState<WorkflowRunModel[]>([]);
  const [tableTotal, setTableTotal] = useState<number>(0);
  const tableColumns: TableProps<WorkflowRunModel>["columns"] = [
    {
      key: "$index",
      align: "center",
      fixed: "left",
      width: 48,
      render: (_, __, index) => (page - 1) * pageSize + index + 1,
    },
    {
      key: "id",
      title: t("workflow_run.props.id"),
      render: (_, record) => <span className="font-mono">{record.id}</span>,
    },
    {
      key: "status",
      title: t("workflow_run.props.status"),
      render: (_, record) => {
        return <WorkflowStatusTag status={record.status} />;
      },
    },
    {
      key: "trigger",
      title: t("workflow_run.props.trigger"),
      ellipsis: true,
      render: (_, record) => {
        if (record.trigger === WORKFLOW_TRIGGERS.SCHEDULED) {
          return t("workflow_run.props.trigger.scheduled");
        } else if (record.trigger === WORKFLOW_TRIGGERS.MANUAL) {
          return t("workflow_run.props.trigger.manual");
        }

        return <></>;
      },
    },
    {
      key: "startedAt",
      title: t("workflow_run.props.started_at"),
      ellipsis: true,
      render: (_, record) => {
        if (record.startedAt) {
          return dayjs(record.startedAt).format("YYYY-MM-DD HH:mm:ss");
        }

        return <></>;
      },
    },
    {
      key: "endedAt",
      title: t("workflow_run.props.ended_at"),
      ellipsis: true,
      render: (_, record) => {
        if (record.endedAt) {
          return dayjs(record.endedAt).format("YYYY-MM-DD HH:mm:ss");
        }

        return <></>;
      },
    },
    {
      key: "$action",
      align: "end",
      fixed: "right",
      width: 120,
      render: (_, record) => {
        const allowCancel = record.status === WORKFLOW_RUN_STATUSES.PENDING || record.status === WORKFLOW_RUN_STATUSES.RUNNING;
        const aloowDelete =
          record.status === WORKFLOW_RUN_STATUSES.SUCCEEDED ||
          record.status === WORKFLOW_RUN_STATUSES.FAILED ||
          record.status === WORKFLOW_RUN_STATUSES.CANCELED;

        return (
          <div className="flex items-center justify-end">
            <Tooltip title={t("workflow_run.action.view.button")}>
              <Button
                color="primary"
                icon={<IconBrowserShare size="1.25em" />}
                variant="text"
                onClick={(e) => {
                  e.stopPropagation();
                  handleRecordDetailClick(record);
                }}
              />
            </Tooltip>
            <Tooltip title={t("workflow_run.action.cancel.button")}>
              <Button
                color="default"
                disabled={!allowCancel}
                icon={<IconPlayerPause size="1.25em" />}
                variant="text"
                onClick={(e) => {
                  e.stopPropagation();
                  handleRecordCancelClick(record);
                }}
              />
            </Tooltip>
            <Tooltip title={t("workflow_run.action.delete.button")}>
              <Button
                color="danger"
                danger
                disabled={!aloowDelete}
                icon={<IconTrash size="1.25em" />}
                variant="text"
                onClick={(e) => {
                  e.stopPropagation();
                  handleRecordDeleteClick(record);
                }}
              />
            </Tooltip>
          </div>
        );
      },
      onCell: () => {
        return {
          onClick: (e) => {
            e.stopPropagation();
          },
        };
      },
    },
  ];

  const {
    loading,
    error: loadedError,
    run: refreshData,
  } = useRequest(
    () => {
      return listWorkflowRuns({
        workflowId: workflow.id,
        page: page,
        perPage: pageSize,
      });
    },
    {
      refreshDeps: [workflow.id, page, pageSize],
      onSuccess: (res) => {
        setTableData(res.items);
        setTableTotal(res.totalItems);
      },
      onError: (err) => {
        if (err instanceof ClientResponseError && err.isAbort) {
          return;
        }

        console.error(err);
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      },
    }
  );

  useEffect(() => {
    const items = tableData.filter((e) => e.status === WORKFLOW_RUN_STATUSES.PENDING || e.status === WORKFLOW_RUN_STATUSES.RUNNING);
    for (const item of items) {
      subscribeWorkflowRun(item.id, (cb) => {
        setTableData((prev) => {
          const index = prev.findIndex((e) => e.id === item.id);
          if (index !== -1) {
            prev[index] = cb.record;
          }
          return [...prev];
        });

        if (cb.record.id === detailDrawerProps.data?.id) {
          setDetailRecord({ ...detailDrawerProps.data, ...cb.record });
        }

        if (cb.record.status !== WORKFLOW_RUN_STATUSES.PENDING && cb.record.status !== WORKFLOW_RUN_STATUSES.RUNNING) {
          unsubscribeWorkflowRun(item.id);
        }
      });
    }

    return () => {
      for (const item of items) {
        unsubscribeWorkflowRun(item.id);
      }
    };
  }, [tableData]);

  const handlePaginationChange = (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
  };

  const { setData: setDetailRecord, setOpen: setDetailOpen, ...detailDrawerProps } = WorkflowRunDetailDrawer.useProps();

  const handleRecordDetailClick = (workflowRun: WorkflowRunModel) => {
    setDetailRecord(workflowRun);
    setDetailOpen(true);
  };

  const handleRecordCancelClick = (workflowRun: WorkflowRunModel) => {
    modal.confirm({
      title: t("workflow_run.action.cancel.modal.title"),
      content: t("workflow_run.action.cancel.modal.content"),
      onOk: async () => {
        try {
          const resp = await cancelWorkflowRun(workflow.id, workflowRun.id);
          if (resp) {
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleRecordDeleteClick = (workflowRun: WorkflowRunModel) => {
    modal.confirm({
      title: <span className="text-error">{t("workflow_run.action.delete.modal.title", { name: `#${workflowRun.id}` })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("workflow_run.action.delete.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeWorkflowRun(workflowRun);
          if (resp) {
            setTableData((prev) => prev.filter((item) => item.id !== workflowRun.id));
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  return (
    <div className="container">
      <div className="pt-9">
        <Alert className="mb-4" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_run.deletion.alert") }}></span>} showIcon type="info" />
        <Alert className="mb-4" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_run.cancellation.alert") }}></span>} showIcon type="info" />

        <Table<WorkflowRunModel>
          columns={tableColumns}
          dataSource={tableData}
          loading={loading}
          locale={{
            emptyText: loading ? (
              <Skeleton />
            ) : (
              <Empty
                title={t("common.text.nodata")}
                description={loadedError ? getErrMsg(loadedError) : t("workflow_run.nodata.description")}
                icon={<IconHistory size={24} />}
              />
            ),
          }}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: tableTotal,
            showSizeChanger: true,
            onChange: handlePaginationChange,
            onShowSizeChange: handlePaginationChange,
          }}
          rowClassName="cursor-pointer"
          rowKey={(record) => record.id}
          scroll={{ x: "max(100%, 960px)" }}
          onRow={(record) => ({
            onClick: () => {
              handleRecordDetailClick(record);
            },
          })}
        />

        <WorkflowRunDetailDrawer {...detailDrawerProps} />
      </div>
    </div>
  );
};

export default WorkflowDetailRuns;
