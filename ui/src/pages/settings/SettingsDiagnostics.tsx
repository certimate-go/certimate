import { useState } from "react";
import { useTranslation } from "react-i18next";
import { IconReload } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { Button, Card, Divider, Empty, List, Pagination, Statistic, Tag, Tooltip, Typography } from "antd";
import dayjs from "dayjs";

import { getStats as getWorkflowStats } from "@/api/workflows";
import Show from "@/components/Show";
import { listCronJobs, listLogs } from "@/repository/system";
import { getNextCronExecutions } from "@/utils/cron";
import { mergeCls } from "@/utils/css";
import { unwrapErrMsg } from "@/utils/error";

const SettingsDiagnostics = () => {
  const { t } = useTranslation();

  return (
    <>
      <h2>{t("settings.diagnostics.logs.title")}</h2>
      <SettingsDiagnosticsLogs />

      <Divider />

      <h2>{t("settings.diagnostics.crons.title")}</h2>
      <SettingsDiagnosticsCrons />

      <Divider />

      <h2>{t("settings.diagnostics.workflow_dispatcher.title")}</h2>
      <SettingsDiagnosticsWorkflowDispatcher />
    </>
  );
};

const SettingsDiagnosticsLogs = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);

  type Log = Awaited<ReturnType<typeof listLogs>>["items"][number];
  const [listData, setListData] = useState<Log[]>([]);

  const [hasMore, setHasMore] = useState(true);

  const {
    loading,
    error: loadError,
    run: refreshData,
  } = useRequest(
    () => {
      return listLogs({ page: page, perPage: pageSize });
    },
    {
      refreshDeps: [page, pageSize],
      debounceWait: 300,
      debounceLeading: true,
      onSuccess: (res) => {
        if (page === 1) {
          setListData([]);
        }

        setListData((prev) => [...prev, ...res.items]);
        setHasMore(res.items.length >= pageSize);
      },
    }
  );

  const renderLogRecord = (record: Log) => {
    let message = <>{record.message}</>;
    if (record.data != null && Object.keys(record.data).length > 0) {
      message = (
        <details>
          <summary>{record.message}</summary>
          {Object.entries(record.data).map(([key, value]) => (
            <div key={key} className="flex space-x-2" style={{ wordBreak: "break-word" }}>
              <div>{key}:</div>
              <div>{JSON.stringify(value)}</div>
            </div>
          ))}
        </details>
      );
    }

    enum LogLevel {
      Info = 0,
      Warn = 4,
      Error = 8,
    }

    return (
      <div className="flex space-x-2">
        <div className="font-mono whitespace-nowrap text-stone-400">[{dayjs(record.created).format("YYYY-MM-DD HH:mm:ss")}]</div>
        <div
          className={mergeCls(
            "flex-1 font-mono",
            +record.level < LogLevel.Info
              ? "text-stone-400"
              : +record.level < LogLevel.Warn
                ? ""
                : +record.level < LogLevel.Error
                  ? "text-warning"
                  : "text-error"
          )}
        >
          {message}
        </div>
      </div>
    );
  };

  const handleReloadClick = () => {
    refreshData();
  };

  const handleRefreshClick = () => {
    setPage(1);
    refreshData();
  };

  const handleLoadMoreClick = () => {
    setPage((prev) => prev + 1);
  };

  return (
    <div className={className} style={style}>
      <div className="size-full overflow-hidden rounded-md bg-black text-stone-200">
        <div className="relative">
          <Show>
            <Show.Case when={loading}>
              <div className="absolute top-4 right-8">
                <Button className="pointer-none" loading>
                  Loading ...
                </Button>
              </div>
            </Show.Case>

            <Show.Case when={listData.length === 0}>
              <div className="px-4 py-2">
                <div className="w-full overflow-hidden">
                  <div className="text-xs leading-relaxed text-stone-400">{loadError ? `> ${unwrapErrMsg(loadError)}` : "> no logs avaiable"}</div>
                </div>
                <div className="flex w-full items-center">
                  <a onClick={handleReloadClick}>
                    <span className="text-xs">{t("common.button.reload")}</span>
                  </a>
                </div>
              </div>
            </Show.Case>

            <Show.Default>
              <div className="px-4 py-2">
                <div className="flex w-full flex-col overflow-hidden">
                  {listData.map((record) => {
                    return (
                      <div key={record.id} className="text-xs leading-relaxed">
                        {renderLogRecord(record)}
                      </div>
                    );
                  })}
                </div>
                <div className="flex w-full items-center">
                  <a onClick={handleRefreshClick}>
                    <span className="text-xs">{t("settings.diagnostics.logs.button.refresh.label")}</span>
                  </a>
                  {hasMore && (
                    <>
                      <Divider orientation="vertical" />
                      <a onClick={handleLoadMoreClick}>
                        <span className="text-xs">{t("settings.diagnostics.logs.button.load_more.label")}</span>
                      </a>
                    </>
                  )}
                </div>
              </div>
            </Show.Default>
          </Show>
        </div>
      </div>
    </div>
  );
};

const SettingsDiagnosticsCrons = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  type CronJob = Awaited<ReturnType<typeof listCronJobs>>["items"][number] & {
    nextTriggerTime: string;
  };
  const [listData, setListData] = useState<CronJob[]>([]);
  const [listTotal, setListTotal] = useState(0);

  const {
    loading,
    error: loadError,
    run: refreshData,
  } = useRequest(
    () => {
      return listCronJobs().then((res) => {
        const startIndex = (page - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        return {
          items: res.items
            .slice(startIndex, endIndex)
            .map((item) => ({
              ...item,
              nextTriggerTime: dayjs(getNextCronExecutions(item.cron)[0]).format("YYYY-MM-DD HH:mm:ss"),
            }))
            .sort((a, b) => a.nextTriggerTime.localeCompare(b.nextTriggerTime)),
          totalItems: res.items.length,
        };
      });
    },
    {
      refreshDeps: [page, pageSize],
      onSuccess: (res) => {
        setListData(res.items);
        setListTotal(res.totalItems);
      },
    }
  );

  const handleReloadClick = () => {
    refreshData();
  };

  const handlePaginationChange = (page: number, pageSize: number) => {
    setPage(page);
    setPageSize(pageSize);
  };

  return (
    <div className={className} style={style}>
      <List<CronJob>
        bordered
        dataSource={listData}
        loading={loading}
        locale={{
          emptyText: (
            <Empty description={loadError ? unwrapErrMsg(loadError) : t("common.text.nodata")} image={Empty.PRESENTED_IMAGE_SIMPLE}>
              {loadError && (
                <Button ghost icon={<IconReload size="1.25em" />} type="primary" onClick={handleReloadClick}>
                  {t("common.button.reload")}
                </Button>
              )}
            </Empty>
          ),
        }}
        rowKey={(record) => record.id}
        renderItem={(record) => (
          <List.Item>
            <Tooltip
              title={
                <>
                  {t("settings.diagnostics.crons.props.next_trigger_time")}
                  <br />
                  {record.nextTriggerTime}
                </>
              }
              mouseEnterDelay={1}
              placement="topRight"
            >
              <div className="flex w-full items-center justify-between gap-4 overflow-hidden xl:hidden">
                <div className="flex-1 truncate">
                  <Typography.Text>{record.id}</Typography.Text>
                </div>
                <div className="text-right">
                  <Typography.Text type="secondary">{record.cron}</Typography.Text>
                </div>
              </div>
            </Tooltip>

            <div className="hidden w-full items-center justify-between gap-4 overflow-hidden xl:flex">
              <div className="flex-1 truncate">
                <Typography.Text>{record.id}</Typography.Text>
              </div>
              <div className="flex items-center justify-end">
                <Typography.Text type="secondary">{record.cron}</Typography.Text>
                <Divider orientation="vertical" />
                <Typography.Text type="secondary">
                  {t("settings.diagnostics.crons.props.next_trigger_time")}
                  {record.nextTriggerTime}
                </Typography.Text>
              </div>
            </div>
          </List.Item>
        )}
      />
      <Show when={page > 1 || listTotal > pageSize}>
        <div className="mt-4 flex justify-end">
          <Pagination current={page} pageSize={pageSize} size="small" total={listTotal} onChange={handlePaginationChange} />
        </div>
      </Show>
    </div>
  );
};

const SettingsDiagnosticsWorkflowDispatcher = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  type Statistics = Awaited<ReturnType<typeof getWorkflowStats>>["data"];
  const [statistics, setStatistics] = useState<Statistics>();

  const { loading, cancel } = useRequest(
    () => {
      return getWorkflowStats();
    },
    {
      throttleWait: 1000,
      throttleLeading: true,
      pollingInterval: 3000,
      pollingWhenHidden: true,
      onSuccess: (res) => {
        setStatistics(res.data);
      },
      onError: () => {
        if (!statistics) {
          cancel();
        }
      },
    }
  );

  return (
    <div className={className} style={style}>
      <div className="flex w-full flex-wrap items-stretch justify-center gap-4 sm:flex-nowrap">
        <Card className="w-full sm:flex-1 md:w-1/3" loading={loading && !statistics}>
          <Statistic title={t("settings.diagnostics.workflow_dispatcher.statistics.concurrency")} value={statistics?.concurrency ?? "-"} />
        </Card>

        <Tooltip
          mouseEnterDelay={1}
          placement="topLeft"
          title={
            statistics?.pendingRunIds?.length
              ? statistics?.pendingRunIds?.map((id) => (
                  <div key={id}>
                    <Tag>#{id}</Tag>
                  </div>
                ))
              : null
          }
        >
          <Card className="w-full sm:flex-1 md:w-1/3" loading={loading && !statistics}>
            <Statistic title={t("settings.diagnostics.workflow_dispatcher.statistics.pending")} value={statistics?.pendingRunIds?.length ?? "-"} />
          </Card>
        </Tooltip>

        <Tooltip
          mouseEnterDelay={1}
          placement="topLeft"
          title={
            statistics?.processingRunIds?.length
              ? statistics?.processingRunIds?.map((id) => (
                  <div key={id}>
                    <Tag>#{id}</Tag>
                  </div>
                ))
              : null
          }
        >
          <Card className="w-full sm:flex-1 md:w-1/3" loading={loading && !statistics}>
            <Statistic title={t("settings.diagnostics.workflow_dispatcher.statistics.processing")} value={statistics?.processingRunIds?.length ?? "-"} />
          </Card>
        </Tooltip>
      </div>
    </div>
  );
};

export default SettingsDiagnostics;
