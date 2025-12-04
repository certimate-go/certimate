import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { IconCirclePlus, IconCopy, IconDots, IconEdit, IconHierarchy3, IconPlayerPlay, IconPlus, IconReload, IconTrash } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Button, Dropdown, Input, Segmented, Skeleton, Switch, Table, type TableProps, Typography, theme } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import { startRun as startWorkflowRun } from "@/api/workflows";
import Empty from "@/components/Empty";
import Show from "@/components/Show";
import WorkflowStatus from "@/components/workflow/WorkflowStatus";
import { WORKFLOW_TRIGGERS, type WorkflowModel, duplicateNodes } from "@/domain/workflow";
import { useAppSettings } from "@/hooks";
import { get as getWorkflow, list as listWorkflows, remove as removeWorkflow, save as saveWorkflow } from "@/repository/workflow";
import { getErrMsg } from "@/utils/error";

const WorkflowList = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { message, modal, notification } = App.useApp();

  const { appSettings: globalAppSettings } = useAppSettings();

  const [filters, setFilters] = useState<Record<string, unknown>>(() => {
    return {
      keyword: searchParams.get("keyword"),
      state: searchParams.get("state"),
    };
  });
  const [sorter, setSorter] = useState<ArrayElement<Parameters<NonNullable<TableProps<WorkflowModel>["onChange"]>>[2]>>(() => {
    return {};
  });
  const [page, setPage] = useState<number>(() => parseInt(+searchParams.get("page")! + "") || 1);
  const [pageSize, setPageSize] = useState<number>(() => parseInt(+searchParams.get("perPage")! + "") || globalAppSettings.defaultPerPage!);

  const [tableData, setTableData] = useState<WorkflowModel[]>([]);
  const [tableTotal, setTableTotal] = useState<number>(0);
  const [tableSelectedRowKeys, setTableSelectedRowKeys] = useState<string[]>([]);
  const tableColumns: TableProps<WorkflowModel>["columns"] = [
    {
      key: "name",
      title: t("workflow.props.name"),
      render: (_, record) => (
        <div className="flex max-w-full flex-col gap-1 truncate">
          <Typography.Text ellipsis>{record.name || "\u00A0"}</Typography.Text>
          <Typography.Text ellipsis type="secondary">
            {record.description || "\u00A0"}
          </Typography.Text>
        </div>
      ),
    },
    {
      key: "trigger",
      title: t("workflow.props.trigger"),
      render: (_, record) => {
        const trigger = record.trigger;
        if (!trigger) {
          return "-";
        } else if (trigger === WORKFLOW_TRIGGERS.MANUAL) {
          return <Typography.Text>{t("workflow.props.trigger.manual")}</Typography.Text>;
        } else if (trigger === WORKFLOW_TRIGGERS.SCHEDULED) {
          return (
            <div className="flex max-w-full flex-col gap-1">
              <Typography.Text>{t("workflow.props.trigger.scheduled")}</Typography.Text>
              <Typography.Text type="secondary">{record.triggerCron || "\u00A0"}</Typography.Text>
            </div>
          );
        }
      },
    },
    {
      key: "state",
      title: t("workflow.props.state"),
      defaultFilteredValue: searchParams.has("state") ? [searchParams.get("state") as string] : void 0,
      render: (_, record) => {
        return (
          <Switch
            checked={record.enabled}
            onChange={() => {
              handleRecordActiveChange(record);
            }}
          />
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
    {
      key: "lastRun",
      title: t("workflow.props.last_run_at"),
      sorter: true,
      sortOrder: sorter.columnKey === "lastRun" ? sorter.order : void 0,
      render: (_, record) => {
        const { lastRunStatus, lastRunTime } = record;
        if (!lastRunStatus) {
          return <></>;
        } else {
          return (
            <WorkflowStatus type="filled" value={lastRunStatus}>
              {lastRunTime ? dayjs(lastRunTime).format("YYYY-MM-DD HH:mm:ss") : ""}
            </WorkflowStatus>
          );
        }
      },
    },
    {
      key: "createdAt",
      title: t("workflow.props.created_at"),
      ellipsis: true,
      render: (_, record) => {
        return dayjs(record.created!).format("YYYY-MM-DD HH:mm:ss");
      },
    },
    {
      key: "$action",
      align: "end",
      fixed: "right",
      width: 64,
      render: (_, record) => (
        <Dropdown
          menu={{
            items: [
              {
                key: "modify",
                label: t("workflow.action.modify.menu"),
                icon: (
                  <span className="anticon scale-125">
                    <IconEdit size="1em" />
                  </span>
                ),
                onClick: () => {
                  handleRecordDetailClick(record);
                },
              },
              {
                key: "duplicate",
                label: t("workflow.action.duplicate.menu"),
                icon: (
                  <span className="anticon scale-125">
                    <IconCopy size="1em" />
                  </span>
                ),
                onClick: () => {
                  handleRecordDuplicateClick(record);
                },
              },
              {
                key: "execute",
                label: t("workflow.action.execute.menu"),
                icon: (
                  <span className="anticon scale-125">
                    <IconPlayerPlay size="1em" />
                  </span>
                ),
                disabled: !record.hasContent,
                onClick: () => {
                  handleRecordExecuteClick(record);
                },
              },
              {
                type: "divider",
              },
              {
                key: "delete",
                label: t("workflow.action.delete.menu"),
                danger: true,
                icon: (
                  <span className="anticon scale-125">
                    <IconTrash size="1em" />
                  </span>
                ),
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
      ),
      onCell: () => {
        return {
          onClick: (e) => {
            e.stopPropagation();
          },
        };
      },
    },
  ];
  const tableRowSelection: TableProps<WorkflowModel>["rowSelection"] = {
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
      const { columnKey: sorterKey, order: sorterOrder } = sorter;
      let sort: string | undefined;
      sort = sorterKey === "lastRun" ? "lastRunTime" : void 0;
      sort = sort && (sorterOrder === "ascend" ? `${sort}` : sorterOrder === "descend" ? `-${sort}` : void 0);

      return listWorkflows({
        keyword: filters["keyword"] as string,
        enabled: (filters["state"] as string) === "enabled" ? true : (filters["state"] as string) === "disabled" ? false : void 0,
        sort: sort,
        page: page,
        perPage: pageSize,
        expand: true,
      });
    },
    {
      refreshDeps: [filters, sorter, page, pageSize],
      onBefore: () => {
        setSearchParams((prev) => {
          if (filters["keyword"]) {
            prev.set("keyword", filters["keyword"] as string);
          } else {
            prev.delete("keyword");
          }

          if (filters["state"]) {
            prev.set("state", filters["state"] as string);
          } else {
            prev.delete("state");
          }

          prev.set("page", page.toString());
          prev.set("perPage", pageSize.toString());

          return prev;
        });
      },
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
        notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      },
    }
  );

  const handleSearch = (value: string) => {
    setFilters((prev) => ({ ...prev, keyword: value.trim() }));
    setPage(1);
  };

  const handleReloadClick = () => {
    if (loading) return;

    refreshData();
  };

  const handlePaginationChange = (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
  };

  const handleCreateClick = () => {
    navigate("/workflows/new");
  };

  const handleRecordDetailClick = (workflow: WorkflowModel) => {
    navigate(`/workflows/${workflow.id}`);
  };

  const handleRecordActiveChange = async (workflow: WorkflowModel) => {
    try {
      if (!workflow.enabled && !workflow.hasContent) {
        message.warning(t("workflow.action.enable.errmsg.unpublished"));
        return;
      }

      const resp = await saveWorkflow({
        id: workflow.id,
        enabled: !tableData.find((item) => item.id === workflow.id)?.enabled,
      });
      if (resp) {
        setTableData((prev) => {
          return prev.map((item) => {
            if (item.id === workflow.id) {
              item.enabled = resp.enabled;
              item.updated = resp.updated;
            }
            return item;
          });
        });
      }
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
    }
  };

  const handleRecordExecuteClick = async (workflow: WorkflowModel) => {
    try {
      await startWorkflowRun(workflow.id);

      message.info(t("workflow.action.execute.prompt"));
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
    }
  };

  const handleRecordDuplicateClick = (workflow: WorkflowModel) => {
    modal.confirm({
      title: t("workflow.action.duplicate.modal.title", { name: workflow.name }),
      content: t("workflow.action.duplicate.modal.content"),
      onOk: async () => {
        try {
          workflow = await getWorkflow(workflow.id);
          const workflowCopy = {
            name: `${workflow.name}-copy`,
            description: workflow.description,
            trigger: workflow.trigger,
            triggerCron: workflow.triggerCron,
            graphDraft: { nodes: duplicateNodes(workflow.graphDraft?.nodes ?? [], { withCopySuffix: false }) },
            hasDraft: true,
          } as WorkflowModel;
          const resp = await saveWorkflow(workflowCopy);
          if (resp) {
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleRecordDeleteClick = (workflow: WorkflowModel) => {
    modal.confirm({
      title: <span className="text-error">{t("workflow.action.delete.modal.title", { name: workflow.name })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("workflow.action.delete.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeWorkflow(workflow);
          if (resp) {
            setTableData((prev) => prev.filter((item) => item.id !== workflow.id));
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleBatchDeleteClick = () => {
    const records = tableData.filter((item) => tableSelectedRowKeys.includes(item.id));
    if (records.length === 0) {
      return;
    }

    modal.confirm({
      title: <span className="text-error">{t("workflow.action.batch_delete.modal.title")}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("workflow.action.batch_delete.modal.content", { count: records.length }) }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeWorkflow(records);
          if (resp) {
            setTableData((prev) => prev.filter((item) => !records.some((record) => record.id === item.id)));
            setTableTotal((prev) => prev - records.length);
            refreshData();
          }
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("workflow.page.title")}</h1>
        <p className="text-base text-gray-500">{t("workflow.page.subtitle")}</p>
      </div>

      <div className="container">
        <div className="flex items-center justify-between gap-x-2 gap-y-3 not-md:flex-col-reverse not-md:items-start not-md:justify-normal">
          <div className="flex w-full flex-1 items-center gap-x-2 md:max-w-200">
            <div>
              <Segmented
                options={[
                  { label: <span className="text-sm">{t("workflow.props.state.filter.all")}</span>, value: "" },
                  { label: <span className="text-sm">{t("workflow.props.state.filter.enabled")}</span>, value: "enabled" },
                  { label: <span className="text-sm">{t("workflow.props.state.filter.disabled")}</span>, value: "disabled" },
                ]}
                size="large"
                value={(filters["state"] as string) || ""}
                onChange={(value) => {
                  setPage(1);
                  setFilters((prev) => ({ ...prev, state: value }));
                }}
              />
            </div>
            <div className="flex-1">
              <Input.Search
                className="text-sm placeholder:text-sm"
                allowClear
                defaultValue={filters["keyword"] as string}
                placeholder={t("workflow.search.placeholder")}
                size="large"
                onSearch={handleSearch}
              />
            </div>
            <div>
              <Button icon={<IconReload size="1.25em" />} size="large" onClick={handleReloadClick} />
            </div>
          </div>
          <div>
            <Button className="text-sm" icon={<IconPlus size="1.25em" />} size="large" type="primary" onClick={handleCreateClick}>
              {t("workflow.action.create.button")}
            </Button>
          </div>
        </div>

        <div className="relative mt-4">
          <Table<WorkflowModel>
            columns={tableColumns}
            dataSource={tableData}
            loading={loading}
            locale={{
              emptyText: loading ? (
                <Skeleton />
              ) : (
                <Empty
                  className="py-24"
                  title={t("workflow.nodata.title")}
                  description={loadError ? getErrMsg(loadError) : t("workflow.nodata.description")}
                  icon={<IconHierarchy3 size={24} />}
                  extra={
                    loadError ? (
                      <Button ghost icon={<IconReload size="1.25em" />} type="primary" onClick={handleReloadClick}>
                        {t("common.button.reload")}
                      </Button>
                    ) : (
                      <Button icon={<IconCirclePlus size="1.25em" />} type="primary" onClick={handleCreateClick}>
                        {t("workflow.nodata.button")}
                      </Button>
                    )
                  }
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
            onChange={(_, __, sorter) => {
              setSorter(Array.isArray(sorter) ? sorter[0] : sorter);
            }}
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
      </div>
    </div>
  );
};

export default WorkflowList;
