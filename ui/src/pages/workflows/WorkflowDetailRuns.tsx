import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconBox, IconBrowserShare, IconCertificate, IconDots, IconHistory, IconPlayerPause, IconTrash } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Button, Dropdown, Skeleton, Table, type TableProps, theme } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import { cancelRun as cancelWorkflowRun } from "@/api/workflows";
import Empty from "@/components/Empty";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import WorkflowRunDetailDrawer from "@/components/workflow/WorkflowRunDetailDrawer";
import WorkflowStatus from "@/components/workflow/WorkflowStatus";
import { WORKFLOW_TRIGGERS } from "@/domain/workflow";
import { WORKFLOW_RUN_STATUSES, type WorkflowRunModel } from "@/domain/workflowRun";
import { useAppSettings, useZustandShallowSelector } from "@/hooks";
import { get as getWorkflowRun, list as listWorkflowRuns, remove as removeWorkflowRun, subscribe as subscribeWorkflowRun } from "@/repository/workflowRun";
import { useWorkflowStore } from "@/stores/workflow";
import { unwrapErrMsg } from "@/utils/error";

const WorkflowDetailRuns = () => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { modal, notification } = App.useApp();

  const { appSettings: globalAppSettings } = useAppSettings();

  const { workflow } = useWorkflowStore(useZustandShallowSelector(["workflow"]));

  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(globalAppSettings.defaultPerPage!);

  const [tableData, setTableData] = useState<WorkflowRunModel[]>([]);
  const [tableTotal, setTableTotal] = useState<number>(0);
  const [tableSelectedRowKeys, setTableSelectedRowKeys] = useState<string[]>([]);
  const tableColumns: TableProps<WorkflowRunModel>["columns"] = [
    {
      key: "id",
      title: "ID",
      width: 160,
      render: (_, record) => <span className="font-mono">{record.id}</span>,
    },
    {
      key: "status",
      title: t("workflow_run.props.status"),
      render: (_, record) => {
        return <WorkflowStatus type="filled" value={record.status} />;
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
      key: "artifacts",
      title: t("workflow_run.props.artifacts"),
      width: 160,
      render: (_, record) => {
        if (record.outputs && record.outputs.length > 0) {
          const keys = new Set<string>();
          const icons: React.ReactNode[] = [];

          for (const output of record.outputs) {
            if (output.type === "ref" && output.value?.split("#")?.at(0) === "certificate") {
              const KEY = "certificate";
              if (keys.has(KEY)) continue;

              keys.add(KEY);
              icons.push(<IconCertificate key={KEY} size="1.25em" />);
            } else {
              const KEY = "other";
              if (keys.has(KEY)) continue;

              keys.add(KEY);
              icons.push(<IconBox key={KEY} size="1.25em" />);
            }
          }

          return <div className="flex items-center gap-2">{icons}</div>;
        }

        return <></>;
      },
    },
    {
      key: "$action",
      align: "end",
      fixed: "right",
      width: 64,
      render: (_, record) => {
        const cancelDisabled = !([WORKFLOW_RUN_STATUSES.PENDING, WORKFLOW_RUN_STATUSES.PROCESSING] as string[]).includes(record.status);
        const deleteDisabled = !cancelDisabled;

        return (
          <Dropdown
            menu={{
              items: [
                {
                  key: "view",
                  label: t("workflow_run.action.view.menu"),
                  icon: (
                    <span className="anticon scale-125">
                      <IconBrowserShare size="1em" />
                    </span>
                  ),
                  onClick: () => {
                    handleRecordDetailClick(record);
                  },
                },
                {
                  key: "cancel",
                  label: <span style={{ color: cancelDisabled ? void 0 : "var(--color-warning)" }}>{t("workflow_run.action.cancel.menu")}</span>,
                  icon: (
                    <span className="anticon scale-125">
                      <IconPlayerPause size="1em" color={cancelDisabled ? void 0 : "var(--color-warning)"} />
                    </span>
                  ),
                  disabled: cancelDisabled,
                  onClick: () => {
                    handleRecordCancelClick(record);
                  },
                },
                {
                  type: "divider",
                },
                {
                  key: "delete",
                  label: t("workflow_run.action.delete.menu"),
                  icon: (
                    <span className="anticon scale-125">
                      <IconTrash size="1em" />
                    </span>
                  ),
                  danger: true,
                  disabled: deleteDisabled,
                  onClick: () => {
                    handleRecordDeleteClick(record);
                  },
                },
              ],
            }}
            trigger={["click"]}
          >
            <Button icon={<IconDots size="1.25em" />} type="text" />
          </Dropdown>
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
  const tableRowSelection: TableProps<WorkflowRunModel>["rowSelection"] = {
    fixed: true,
    selectedRowKeys: tableSelectedRowKeys,
    renderCell(checked, _, index, node) {
      if (!checked) {
        return (
          <div className="group/selection">
            <div className="group-hover/selection:hidden">{(page - 1) * pageSize + index + 1}</div>
            <div className="hidden group-hover/selection:block">{node}</div>
          </div>
        );
      }
      return node;
    },
    onCell: () => {
      return {
        onClick: (e) => {
          e.stopPropagation();
        },
      };
    },
    onChange: (keys) => {
      setTableSelectedRowKeys(keys as string[]);
    },
  };

  const {
    loading,
    error: loadError,
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
        setTableSelectedRowKeys([]);
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

  useEffect(() => {
    const unsubscribers = new Map<string, () => void>();
    const items = tableData.filter((e) => e.status === WORKFLOW_RUN_STATUSES.PENDING || e.status === WORKFLOW_RUN_STATUSES.PROCESSING);
    for (const item of items) {
      subscribeWorkflowRun(item.id, (cb) => {
        setTableData((prev) => {
          const index = prev.findIndex((e) => e.id === item.id);
          if (index !== -1) {
            prev[index] = cb.record;
          }
          return [...prev];
        });

        if (cb.record.status !== WORKFLOW_RUN_STATUSES.PENDING && cb.record.status !== WORKFLOW_RUN_STATUSES.PROCESSING) {
          unsubscribers.get(cb.record.id)?.();
          unsubscribers.delete(cb.record.id);
        }
      }).then((unsubscriber) => {
        unsubscribers.set(item.id, unsubscriber);
      });
    }

    return () => {
      for (const unsubscriber of unsubscribers.values()) {
        unsubscriber();
      }
      unsubscribers.clear();
    };
  }, [tableData]);

  const handlePaginationChange = (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
  };

  const { drawerProps: detailDrawerProps, ...detailDrawer } = WorkflowRunDetailDrawer.useDrawer();

  const handleRecordDetailClick = (workflowRun: WorkflowRunModel) => {
    const drawer = detailDrawer.open({ data: workflowRun, loading: true });
    getWorkflowRun(workflowRun.id).then((data) => {
      drawer.safeUpdate({ data, loading: false });
    });
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
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
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
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  const handleBatchDeleteClick = () => {
    let records = tableData.filter((item) => tableSelectedRowKeys.includes(item.id));
    if (records.length === 0) {
      return;
    }

    modal.confirm({
      title: <span className="text-error">{t("workflow_run.action.batch_delete.modal.title")}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("workflow_run.action.batch_delete.modal.content", { count: records.length }) }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        // 未结束的记录不允许删除
        records = records.filter((record) => !([WORKFLOW_RUN_STATUSES.PENDING, WORKFLOW_RUN_STATUSES.PROCESSING] as string[]).includes(record.status));
        try {
          const resp = await removeWorkflowRun(records);
          if (resp) {
            setTableData((prev) => prev.filter((item) => !records.some((record) => record.id === item.id)));
            setTableTotal((prev) => prev - records.length);
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
        }
      },
    });
  };

  return (
    <div className="container">
      <div className="pt-9">
        <Tips className="mb-4" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_run.deletion.alert") }}></span>} />
        <Tips className="mb-4" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_run.cancellation.alert") }}></span>} />

        <div className="relative">
          <Table<WorkflowRunModel>
            columns={tableColumns}
            dataSource={tableData}
            loading={loading}
            locale={{
              emptyText: loading ? (
                <Skeleton />
              ) : (
                <Empty
                  className="py-24"
                  title={loadError ? t("common.text.nodata_failed") : t("workflow_run.nodata.title")}
                  description={loadError ? unwrapErrMsg(loadError) : t("workflow_run.nodata.description")}
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
            rowSelection={tableRowSelection}
            scroll={{ x: "max(100%, 960px)" }}
            onRow={(record) => ({
              onClick: () => {
                handleRecordDetailClick(record);
              },
            })}
          />

          <Show when={tableSelectedRowKeys.length > 0}>
            <div
              className="absolute top-0 right-0 left-[32px] z-10 h-[54px]"
              style={{
                left: "32px", // Match the width of the table row selection checkbox
                height: "54px", // Match the height of the table header
                background: themeToken.Table?.headerBg ?? themeToken.colorBgElevated,
              }}
            >
              <div className="flex size-full items-center justify-end gap-x-2 overflow-hidden px-4 py-2">
                <Button danger ghost onClick={handleBatchDeleteClick}>
                  {t("common.button.delete")}
                </Button>
              </div>
            </div>
          </Show>
        </div>

        <WorkflowRunDetailDrawer {...detailDrawerProps} />
      </div>
    </div>
  );
};

export default WorkflowDetailRuns;
