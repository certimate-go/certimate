import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { InputNumber, Select, Typography } from "antd";

import { type WorkflowSimpleSchedule } from "@/domain/workflow";

export interface SimplifiedScheduleInputProps {
  value?: WorkflowSimpleSchedule;
  onChange?: (value: WorkflowSimpleSchedule) => void;
}

const MAX_TIME_POINTS = 10;
const HOUR_OPTIONS = Array.from({ length: 24 }, (_, hour) => {
  const timePoint = `${String(hour).padStart(2, "0")}:00`;
  return {
    label: timePoint,
    value: timePoint,
  };
});
const DEFAULT_VALUE: WorkflowSimpleSchedule = {
  intervalDays: 1,
  timePoints: ["00:00"],
};

const SimplifiedScheduleInput = ({ value, onChange }: SimplifiedScheduleInputProps) => {
  const { t } = useTranslation();
  const schedule = value ?? DEFAULT_VALUE;
  const [message, setMessage] = useState<string>();

  const sortedTimePoints = useMemo(() => sortTimePoints(schedule.timePoints), [schedule.timePoints]);
  const maxReached = sortedTimePoints.length >= MAX_TIME_POINTS;

  const emitChange = (next: WorkflowSimpleSchedule) => {
    onChange?.({
      intervalDays: next.intervalDays,
      timePoints: sortTimePoints(next.timePoints),
    });
  };

  const handleIntervalDaysChange = (nextValue: number | null) => {
    emitChange({
      ...schedule,
      intervalDays: nextValue ?? 1,
    });
  };

  const handleTimePointsChange = (nextTimePoints: string[]) => {
    if (nextTimePoints.length > MAX_TIME_POINTS) {
      setMessage(t("workflow_node.start.form.schedule.errmsg.max_time"));
      return;
    }

    setMessage(undefined);
    emitChange({
      ...schedule,
      timePoints: nextTimePoints,
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <Typography.Text>{t("workflow_node.start.form.schedule.interval_days")}</Typography.Text>
        <InputNumber className="w-24" min={1} max={31} value={schedule.intervalDays} onChange={handleIntervalDaysChange} />
        <Typography.Text>{t("workflow_node.start.form.schedule.interval_days_suffix")}</Typography.Text>
      </div>

      <div className="space-y-2">
        <Typography.Text className="block">{t("workflow_node.start.form.schedule.time_points")}</Typography.Text>
        <Select
          className="w-full max-w-80"
          allowClear
          mode="multiple"
          maxTagCount="responsive"
          options={HOUR_OPTIONS}
          placeholder={t("workflow_node.start.form.schedule.hour_placeholder")}
          value={sortedTimePoints}
          onChange={handleTimePointsChange}
        />

        {sortedTimePoints.length === 0 ? <Typography.Text type="danger">{t("workflow_node.start.form.schedule.errmsg.no_time")}</Typography.Text> : null}
        {message ? <Typography.Text type="danger">{message}</Typography.Text> : null}
        {maxReached ? <Typography.Text type="secondary">{t("workflow_node.start.form.schedule.errmsg.max_time")}</Typography.Text> : null}
      </div>
    </div>
  );
};

function sortTimePoints(timePoints: string[]): string[] {
  return [...new Set(timePoints)].sort((a, b) => timePointToMinutes(a) - timePointToMinutes(b));
}

function timePointToMinutes(timePoint: string): number {
  const [hour, minute] = timePoint.split(":").map((part) => parseInt(part, 10));
  return hour * 60 + minute;
}

export default SimplifiedScheduleInput;
