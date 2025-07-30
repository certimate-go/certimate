import { useTranslation } from "react-i18next";
import { Tag } from "antd";

import { WORKFLOW_RUN_STATUSES } from "@/domain/workflowRun";

import WorkflowStatusIcon from "./WorkflowStatusIcon";

export interface WorkflowStatusTagProps {
  className?: string;
  style?: React.CSSProperties;
  status: string;
}

const WorkflowStatusTag = ({ className, style, status }: WorkflowStatusTagProps) => {
  const { t } = useTranslation();

  const Icon = <WorkflowStatusIcon status={status} />;

  switch (status) {
    case WORKFLOW_RUN_STATUSES.PENDING:
      return (
        <Tag className={className} style={style} icon={Icon}>
          {t("workflow_run.props.status.pending")}
        </Tag>
      );
    case WORKFLOW_RUN_STATUSES.RUNNING:
      return (
        <Tag className={className} style={style} color="var(--color-info)" icon={Icon}>
          {t("workflow_run.props.status.running")}
        </Tag>
      );
    case WORKFLOW_RUN_STATUSES.SUCCEEDED:
      return (
        <Tag className={className} style={style} color="var(--color-success)" icon={Icon}>
          {t("workflow_run.props.status.succeeded")}
        </Tag>
      );
    case WORKFLOW_RUN_STATUSES.FAILED:
      return (
        <Tag className={className} style={style} color="var(--color-error)" icon={Icon}>
          {t("workflow_run.props.status.failed")}
        </Tag>
      );
    case WORKFLOW_RUN_STATUSES.CANCELED:
      return (
        <Tag className={className} style={style} color="var(--color-warning)" icon={Icon}>
          {t("workflow_run.props.status.canceled")}
        </Tag>
      );
    default:
      return <></>;
  }
};

export default WorkflowStatusTag;
