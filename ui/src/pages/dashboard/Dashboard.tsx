import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import {
  IconActivity,
  IconAlertHexagon,
  IconCirclePlus,
  IconExternalLink,
  IconHexagonLetterX,
  IconHistory,
  IconLock,
  IconPlugConnected,
  IconReload,
  IconRoute,
  IconShieldCheckered,
} from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Button, Card, Col, Divider, Flex, Grid, Row, Skeleton, Table, type TableProps, Typography } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import { get as getStatistics } from "@/api/statistics";
import Empty from "@/components/Empty";
import WorkflowRunDetailDrawer from "@/components/workflow/WorkflowRunDetailDrawer";
import WorkflowStatusTag from "@/components/workflow/WorkflowStatusTag";
import { type Statistics } from "@/domain/statistics";
import { type WorkflowRunModel } from "@/domain/workflowRun";
import { useBrowserTheme } from "@/hooks";
import { list as listWorkflowRuns } from "@/repository/workflowRun";
import { mergeCls } from "@/utils/css";
import { getErrMsg } from "@/utils/error";

const Dashboard = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const breakpoints = Grid.useBreakpoint();

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("dashboard.page.title")}</h1>
      </div>

      <div className="container">
        <div className="my-[6px]">
          <StatisticCards />
        </div>

        <Divider />

        <Flex justify="stretch" vertical={!breakpoints.lg} gap={16}>
          <Card className="transition-all max-lg:flex-1 lg:w-[280px] xl:w-[360px]" title={t("dashboard.quick_actions")}>
            <div className="flex flex-col gap-4">
              <Button block type="primary" size="large" icon={<IconCirclePlus size="1.25em" />} onClick={() => navigate("/workflows/new")}>
                <span className="text-sm">{t("dashboard.quick_actions.create_workflow")}</span>
              </Button>
              <Button block size="large" icon={<IconLock size="1.25em" />} onClick={() => navigate("/settings/account")}>
                <span className="text-sm">{t("dashboard.quick_actions.change_password")}</span>
              </Button>
              <Button block size="large" icon={<IconPlugConnected size="1.25em" />} onClick={() => navigate("/settings/ssl-provider")}>
                <span className="text-sm">{t("dashboard.quick_actions.configure_ca")}</span>
              </Button>
            </div>
          </Card>
          <Card className="flex-1" title={t("dashboard.latest_workflow_runs")}>
            <WorkflowRunHistoryTable />
          </Card>
        </Flex>
      </div>
    </div>
  );
};

const StatisticCard = ({
  className,
  style,
  label,
  loading,
  icon,
  value,
  onClick,
}: {
  className?: string;
  style?: React.CSSProperties;
  label: React.ReactNode;
  loading?: boolean;
  icon: React.ReactNode;
  value?: string | number | React.ReactNode;
  onClick?: () => void;
}) => {
  return (
    <Card
      className={mergeCls("size-full overflow-hidden ", className)}
      style={style}
      styles={{ body: { padding: 0 } }}
      hoverable
      loading={loading}
      variant="borderless"
      onClick={onClick}
    >
      <div className="relative overflow-hidden pt-6 pr-4 pb-4 pl-6">
        <div className="absolute inset-0 z-0 bg-stone-200 opacity-10">
          <div
            className="size-full"
            style={{
              backgroundImage:
                "linear-gradient(rgba(255, 255, 255, 0.8) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.8) 1px, transparent 1px)",
              backgroundSize: "20px 20px",
            }}
          />
        </div>
        <div className="mb-2">
          <div className="truncate text-sm font-medium text-white/75">{label}</div>
        </div>
        <div className="relative flex items-center justify-between">
          <div className="truncate text-4xl font-medium text-white">{value}</div>
          <div className="flex size-12 items-center justify-center rounded-full bg-white/25 p-3 text-white/75">{icon}</div>
        </div>
      </div>
    </Card>
  );
};

const StatisticCards = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { theme: browserTheme } = useBrowserTheme();

  const { notification } = App.useApp();

  const cardGridSpans = {
    xs: { flex: "50%" },
    md: { flex: "50%" },
    lg: { flex: "33.3333%" },
    xl: { flex: "33.3333%" },
    xxl: { flex: "20%" },
  };
  const cardStylesFn = (color: string) => ({
    background:
      browserTheme === "dark"
        ? `linear-gradient(135deg, color-mix(in srgb, ${color} 50%, black 20%) 0%, color-mix(in srgb, ${color} 50%, white 20%) 100%)`
        : `linear-gradient(135deg, color-mix(in srgb, ${color} 80%, black 30%) 0%, color-mix(in srgb, ${color} 80%, white 30%) 100%)`,
  });

  const [statistics, setStatistics] = useState<Statistics>();

  const { loading } = useRequest(
    () => {
      return getStatistics();
    },
    {
      onSuccess: (res) => {
        setStatistics(res.data);
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

  return (
    <div className={className} style={style}>
      <Row className="justify-stretch" gutter={[16, 16]}>
        <Col className="overflow-hidden" {...cardGridSpans}>
          <StatisticCard
            style={cardStylesFn("var(--color-info)")}
            icon={<IconShieldCheckered size={48} />}
            label={t("dashboard.statistics.all_certificates")}
            loading={loading}
            value={statistics?.certificateTotal ?? "-"}
            onClick={() => navigate("/certificates")}
          />
        </Col>
        <Col className="overflow-hidden" {...cardGridSpans}>
          <StatisticCard
            style={cardStylesFn("var(--color-warning)")}
            icon={<IconAlertHexagon size={48} />}
            label={t("dashboard.statistics.expire_soon_certificates")}
            loading={loading}
            value={statistics?.certificateExpireSoon ?? "-"}
            onClick={() => navigate("/certificates?state=expireSoon")}
          />
        </Col>
        <Col className="overflow-hidden" {...cardGridSpans}>
          <StatisticCard
            style={cardStylesFn("var(--color-error)")}
            icon={<IconHexagonLetterX size={48} />}
            label={t("dashboard.statistics.expired_certificates")}
            loading={loading}
            value={statistics?.certificateExpired ?? "-"}
            onClick={() => navigate("/certificates?state=expired")}
          />
        </Col>
        <Col className="overflow-hidden" {...cardGridSpans}>
          <StatisticCard
            style={cardStylesFn("var(--color-info)")}
            icon={<IconRoute size={48} />}
            label={t("dashboard.statistics.all_workflows")}
            loading={loading}
            value={statistics?.workflowTotal ?? "-"}
            onClick={() => navigate("/workflows")}
          />
        </Col>
        <Col className="overflow-hidden" {...cardGridSpans}>
          <StatisticCard
            style={cardStylesFn("var(--color-success)")}
            icon={<IconActivity size={48} />}
            label={t("dashboard.statistics.enabled_workflows")}
            loading={loading}
            value={statistics?.workflowEnabled ?? "-"}
            onClick={() => navigate("/workflows?state=enabled")}
          />
        </Col>
      </Row>
    </div>
  );
};

const WorkflowRunHistoryTable = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const [tableData, setTableData] = useState<WorkflowRunModel[]>([]);
  const tableColumns: TableProps<WorkflowRunModel>["columns"] = [
    {
      key: "$index",
      align: "center",
      fixed: "left",
      width: 48,
      render: (_, __, index) => index + 1,
    },
    {
      key: "name",
      title: t("workflow.props.name"),
      render: (_, record) => {
        const workflow = record.expand?.workflowRef;
        return (
          <div className="max-w-full truncate">
            <Typography.Link
              ellipsis
              onClick={() => {
                if (workflow) {
                  navigate(`/workflows/${workflow.id}`);
                }
              }}
            >
              {workflow?.name ?? <span className="font-mono">{t(`#${record.workflowRef}`)}</span>}
            </Typography.Link>
          </div>
        );
      },
    },
    {
      key: "status",
      title: t("workflow_run.props.status"),
      render: (_, record) => {
        return <WorkflowStatusTag status={record.status} />;
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
  ];

  const {
    loading,
    error: loadedError,
    run: refreshData,
  } = useRequest(
    () => {
      return listWorkflowRuns({
        page: 1,
        perPage: 15,
        expand: true,
      });
    },
    {
      onSuccess: (res) => {
        setTableData(res.items);
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

  const handleReloadClick = () => {
    if (loading) return;

    refreshData();
  };

  const { setData: setDetailRecord, setOpen: setDetailOpen, ...detailDrawerProps } = WorkflowRunDetailDrawer.useProps();

  const handleRecordDetailClick = (workflowRun: WorkflowRunModel) => {
    setDetailRecord(workflowRun);
    setDetailOpen(true);
  };

  return (
    <div className={className} style={style}>
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
              description={loadedError ? getErrMsg(loadedError) : t("dashboard.latest_workflow_runs.nodata.description")}
              icon={<IconHistory size={24} />}
              extra={
                loadedError ? (
                  <Button icon={<IconReload size="1.25em" />} type="primary" onClick={handleReloadClick}>
                    {t("common.button.reload")}
                  </Button>
                ) : (
                  <Button icon={<IconExternalLink size="1.25em" />} type="primary" onClick={() => navigate("/workflows")}>
                    {t("dashboard.latest_workflow_runs.nodata.button")}
                  </Button>
                )
              }
            />
          ),
        }}
        pagination={false}
        rowClassName="cursor-pointer"
        rowKey={(record) => record.id}
        scroll={{ x: "max(100%, 720px)" }}
        size="small"
        onRow={(record) => ({
          onClick: () => {
            handleRecordDetailClick(record);
          },
        })}
      />

      <WorkflowRunDetailDrawer {...detailDrawerProps} />
    </div>
  );
};

export default Dashboard;
