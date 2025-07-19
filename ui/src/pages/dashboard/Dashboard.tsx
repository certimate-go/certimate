import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import {
  CheckCircleOutlined as CheckCircleOutlinedIcon,
  ClockCircleOutlined as ClockCircleOutlinedIcon,
  CloseCircleOutlined as CloseCircleOutlinedIcon,
  StopOutlined as StopOutlinedIcon,
  SyncOutlined as SyncOutlinedIcon,
} from "@ant-design/icons";
import {
  IconActivity,
  IconBrowserShare,
  IconPlugConnected,
  IconPlus,
  IconSchema,
  IconShieldCheckered,
  IconShieldExclamation,
  IconShieldX,
  IconUserShield,
} from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { App, Button, Card, Col, Divider, Empty, Flex, Grid, Row, Space, Statistic, Table, type TableProps, Tag, Typography, theme } from "antd";
import dayjs from "dayjs";
import { ClientResponseError } from "pocketbase";

import { get as getStatistics } from "@/api/statistics";
import WorkflowRunDetailDrawer from "@/components/workflow/WorkflowRunDetailDrawer";
import { type Statistics } from "@/domain/statistics";
import { WORKFLOW_RUN_STATUSES, type WorkflowRunModel } from "@/domain/workflowRun";
import { list as listWorkflowRuns } from "@/repository/workflowRun";
import { getErrMsg } from "@/utils/error";

const Dashboard = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();
  const { token: themeToken } = theme.useToken();
  const breakpoints = Grid.useBreakpoint();

  const statisticsGridSpans = {
    xs: { flex: "50%" },
    md: { flex: "50%" },
    lg: { flex: "33.3333%" },
    xl: { flex: "33.3333%" },
    xxl: { flex: "20%" },
  };
  const [statistics, setStatistics] = useState<Statistics>();
  const { loading: statisticsLoading } = useRequest(
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

  const tableColumns: TableProps<WorkflowRunModel>["columns"] = [
    {
      key: "$index",
      align: "center",
      fixed: "left",
      width: 50,
      render: (_, __, index) => index + 1,
    },
    {
      key: "name",
      title: t("workflow.props.name"),
      ellipsis: true,
      render: (_, record) => {
        const workflow = record.expand?.workflowId;
        return (
          <Typography.Link
            ellipsis
            onClick={() => {
              if (workflow) {
                navigate(`/workflows/${workflow.id}`);
              }
            }}
          >
            {workflow?.name ?? <span className="font-mono">{t(`#${record.workflowId}`)}</span>}
          </Typography.Link>
        );
      },
    },
    {
      key: "status",
      title: t("workflow_run.props.status"),
      ellipsis: true,
      render: (_, record) => {
        if (record.status === WORKFLOW_RUN_STATUSES.PENDING) {
          return <Tag icon={<ClockCircleOutlinedIcon />}>{t("workflow_run.props.status.pending")}</Tag>;
        } else if (record.status === WORKFLOW_RUN_STATUSES.RUNNING) {
          return (
            <Tag icon={<SyncOutlinedIcon spin />} color="processing">
              {t("workflow_run.props.status.running")}
            </Tag>
          );
        } else if (record.status === WORKFLOW_RUN_STATUSES.SUCCEEDED) {
          return (
            <Tag icon={<CheckCircleOutlinedIcon />} color="success">
              {t("workflow_run.props.status.succeeded")}
            </Tag>
          );
        } else if (record.status === WORKFLOW_RUN_STATUSES.FAILED) {
          return (
            <Tag icon={<CloseCircleOutlinedIcon />} color="error">
              {t("workflow_run.props.status.failed")}
            </Tag>
          );
        } else if (record.status === WORKFLOW_RUN_STATUSES.CANCELED) {
          return (
            <Tag icon={<StopOutlinedIcon />} color="warning">
              {t("workflow_run.props.status.canceled")}
            </Tag>
          );
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
      width: 120,
      render: (_, record) => (
        <Space.Compact>
          <WorkflowRunDetailDrawer data={record} trigger={<Button color="primary" icon={<IconBrowserShare size="1.25em" />} variant="text" />} />
        </Space.Compact>
      ),
    },
  ];
  const [tableData, setTableData] = useState<WorkflowRunModel[]>([]);
  const { loading: tableLoading } = useRequest(
    () => {
      return listWorkflowRuns({
        page: 1,
        perPage: 9,
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

  return (
    <div className="px-6 py-4">
      <div className="mx-auto max-w-320">
        <h1>{t("dashboard.page.title")}</h1>

        <Row className="justify-stretch" gutter={[16, 16]}>
          <Col {...statisticsGridSpans}>
            <StatisticCard
              icon={<IconShieldCheckered size={48} strokeWidth={1} color={themeToken.colorInfo} />}
              label={t("dashboard.statistics.all_certificates")}
              loading={statisticsLoading}
              value={statistics?.certificateTotal ?? "-"}
              suffix={t("dashboard.statistics.unit")}
              onClick={() => navigate("/certificates")}
            />
          </Col>
          <Col {...statisticsGridSpans}>
            <StatisticCard
              icon={<IconShieldExclamation size={48} strokeWidth={1} color={themeToken.colorWarning} />}
              label={t("dashboard.statistics.expire_soon_certificates")}
              loading={statisticsLoading}
              value={statistics?.certificateExpireSoon ?? "-"}
              suffix={t("dashboard.statistics.unit")}
              onClick={() => navigate("/certificates?state=expireSoon")}
            />
          </Col>
          <Col {...statisticsGridSpans}>
            <StatisticCard
              icon={<IconShieldX size={48} strokeWidth={1} color={themeToken.colorError} />}
              label={t("dashboard.statistics.expired_certificates")}
              loading={statisticsLoading}
              value={statistics?.certificateExpired ?? "-"}
              suffix={t("dashboard.statistics.unit")}
              onClick={() => navigate("/certificates?state=expired")}
            />
          </Col>
          <Col {...statisticsGridSpans}>
            <StatisticCard
              icon={<IconSchema size={48} strokeWidth={1} color={themeToken.colorInfo} />}
              label={t("dashboard.statistics.all_workflows")}
              loading={statisticsLoading}
              value={statistics?.workflowTotal ?? "-"}
              suffix={t("dashboard.statistics.unit")}
              onClick={() => navigate("/workflows")}
            />
          </Col>
          <Col {...statisticsGridSpans}>
            <StatisticCard
              icon={<IconActivity size={48} strokeWidth={1} color={themeToken.colorSuccess} />}
              label={t("dashboard.statistics.enabled_workflows")}
              loading={statisticsLoading}
              value={statistics?.workflowEnabled ?? "-"}
              suffix={t("dashboard.statistics.unit")}
              onClick={() => navigate("/workflows?state=enabled")}
            />
          </Col>
        </Row>

        <Divider />

        <Flex justify="stretch" vertical={!breakpoints.lg} gap={16}>
          <Card className="max-lg:flex-1 lg:w-[360px]" title={t("dashboard.quick_actions")}>
            <Space className="w-full" direction="vertical" size="large">
              <Button block type="primary" size="large" icon={<IconPlus size="1.25em" />} onClick={() => navigate("/workflows/new")}>
                {t("dashboard.quick_actions.create_workflow")}
              </Button>
              <Button block size="large" icon={<IconUserShield size="1.25em" />} onClick={() => navigate("/settings/account")}>
                {t("dashboard.quick_actions.change_login_password")}
              </Button>
              <Button block size="large" icon={<IconPlugConnected size="1.25em" />} onClick={() => navigate("/settings/ssl-provider")}>
                {t("dashboard.quick_actions.configure_ca")}
              </Button>
            </Space>
          </Card>
          <Card className="flex-1" title={t("dashboard.latest_workflow_runs")}>
            <Table<WorkflowRunModel>
              columns={tableColumns}
              dataSource={tableData}
              loading={tableLoading}
              locale={{
                emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />,
              }}
              pagination={false}
              rowKey={(record) => record.id}
              scroll={{ x: "max(100%, 720px)" }}
              size="small"
            />
          </Card>
        </Flex>
      </div>
    </div>
  );
};

const StatisticCard = ({
  label,
  loading,
  icon,
  value,
  suffix,
  onClick,
}: {
  label: React.ReactNode;
  loading?: boolean;
  icon: React.ReactNode;
  value?: string | number | React.ReactNode;
  suffix?: React.ReactNode;
  onClick?: () => void;
}) => {
  return (
    <Card className="size-full overflow-hidden" hoverable loading={loading} variant="borderless" onClick={onClick}>
      <div className="flex gap-2">
        {icon}
        <Statistic
          title={label}
          valueRender={() => {
            return <Typography.Text className="text-4xl">{value}</Typography.Text>;
          }}
          suffix={<Typography.Text className="text-sm">{suffix}</Typography.Text>}
        />
      </div>
    </Card>
  );
};

export default Dashboard;
