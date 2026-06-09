import { CronExpressionParser } from "cron-parser";

import { type WorkflowSimpleSchedule } from "@/domain/workflow";

export const validateCronExpression = (expr: string): boolean => {
  try {
    CronExpressionParser.parse(expr);

    // pocketbase 后端仅支持标准 crontab 形式的表达式
    // 这里转译了来自 pocketbase 的 golang 代码来验证
    const segments = expr.trim().split(" ");
    if (segments.length !== 5) return false;
    parseCronSegment(segments[0], 0, 59);
    parseCronSegment(segments[1], 0, 23);
    parseCronSegment(segments[2], 1, 31);
    parseCronSegment(segments[3], 1, 12);
    parseCronSegment(segments[4], 0, 6);

    return true;
  } catch {
    return false;
  }
};

export const getNextCronExecutions = (expr: string, times = 1): Date[] => {
  if (!validateCronExpression(expr)) return [];

  const now = new Date();
  const cron = CronExpressionParser.parse(expr, { currentDate: now });

  return cron.take(times).map((date) => date.toDate());
};

export const simpleScheduleToCron = (schedule: WorkflowSimpleSchedule): string => {
  if (!Number.isInteger(schedule.intervalDays) || schedule.intervalDays < 1 || schedule.intervalDays > 31) {
    throw new Error("Invalid interval days");
  }

  const timePoints = normalizeTimePoints(schedule.timePoints);
  if (timePoints.length === 0) {
    throw new Error("At least one time point is required");
  }

  const minute = timePoints[0].minute;
  if (minute !== 0 || timePoints.some((timePoint) => timePoint.minute !== 0)) {
    throw new Error("Simple schedules only support hourly time points");
  }

  const hourSegment = timePoints.map((timePoint) => timePoint.hour).join(",");
  const dayOfMonthSegment = schedule.intervalDays === 1 ? "*" : `*/${schedule.intervalDays}`;

  return `0 ${hourSegment} ${dayOfMonthSegment} * *`;
};

export const cronToSimpleSchedule = (expr: string): WorkflowSimpleSchedule | null => {
  try {
    if (!validateCronExpression(expr)) return null;

    const segments = expr.trim().split(/\s+/);
    if (segments.length !== 5) return null;

    const [minuteSegment, hourSegment, dayOfMonthSegment, monthSegment, dayOfWeekSegment] = segments;
    if (monthSegment !== "*" || dayOfWeekSegment !== "*") return null;

    const intervalDays = parseSimpleIntervalDays(dayOfMonthSegment);
    if (intervalDays == null) return null;

    const minutes = parseSingleValueList(minuteSegment, 0, 59);
    const hours = parseSingleValueList(hourSegment, 0, 23);
    if (minutes.length !== 1 || minutes[0] !== 0 || hours.length === 0) return null;

    return {
      intervalDays,
      timePoints: hours.map((hour) => formatTimePoint(hour, minutes[0])),
    };
  } catch {
    return null;
  }
};

export const isSimpleCronPattern = (expr: string): boolean => {
  return cronToSimpleSchedule(expr) != null;
};

// transpile from:
//   https://github.com/pocketbase/pocketbase/blob/5d964c1b1d020f425299b32df03ecf44e0a0502e/tools/cron/schedule.go#L141-L218
function parseCronSegment(segment: string, min: number, max: number): Set<number> {
  const slots = new Set<number>();

  const list = segment.split(",");
  for (const p of list) {
    const stepParts = p.split("/");

    let step: number;
    switch (stepParts.length) {
      case 1:
        {
          step = 1;
        }
        break;
      case 2:
        {
          const parsedStep = parseInt(stepParts[1], 10);
          if (isNaN(parsedStep) || parsedStep < 1 || parsedStep > max) {
            throw new Error(`Invalid segment step boundary - the step must be between 1 and the ${max}`);
          }
          step = parsedStep;
        }
        break;

      default:
        throw new Error("Invalid segment step format - must be in the format */n or 1-30/n");
    }

    let rangeMin: number, rangeMax: number;
    if (stepParts[0] === "*") {
      rangeMin = min;
      rangeMax = max;
    } else {
      const rangeParts = stepParts[0].split("-");
      switch (rangeParts.length) {
        case 1:
          {
            if (step !== 1) {
              throw new Error("Invalid segment step - step > 1 could be used only with the wildcard or range format");
            }
            const parsed = parseInt(rangeParts[0], 10);
            if (isNaN(parsed) || parsed < min || parsed > max) {
              throw new Error("Invalid segment value - must be between the min and max of the segment");
            }
            rangeMin = parsed;
            rangeMax = rangeMin;
          }
          break;

        case 2:
          {
            const parsedMin = parseInt(rangeParts[0], 10);
            if (isNaN(parsedMin) || parsedMin < min || parsedMin > max) {
              throw new Error(`Invalid segment range minimum - must be between ${min} and ${max}`);
            }
            rangeMin = parsedMin;

            const parsedMax = parseInt(rangeParts[1], 10);
            if (isNaN(parsedMax) || parsedMax < rangeMin || parsedMax > max) {
              throw new Error(`Invalid segment range maximum - must be between ${rangeMin} and ${max}`);
            }
            rangeMax = parsedMax;
          }
          break;

        default:
          throw new Error("Invalid segment range format - the range must have 1 or 2 parts");
      }
    }

    for (let i = rangeMin; i <= rangeMax; i += step) {
      slots.add(i);
    }
  }

  return slots;
}

function parseSimpleIntervalDays(segment: string): number | null {
  if (segment === "*") return 1;

  const matches = segment.match(/^\*\/([1-9]\d*)$/);
  if (!matches) return null;

  const intervalDays = parseInt(matches[1], 10);
  return intervalDays >= 1 && intervalDays <= 31 ? intervalDays : null;
}

function parseSingleValueList(segment: string, min: number, max: number): number[] {
  const values = segment.split(",").map((part) => {
    if (!/^\d+$/.test(part)) {
      throw new Error("Invalid simple cron segment");
    }

    const value = parseInt(part, 10);
    if (value < min || value > max) {
      throw new Error("Invalid simple cron segment value");
    }

    return value;
  });

  return [...new Set(values)].sort((a, b) => a - b);
}

function normalizeTimePoints(timePoints: string[]): Array<{ hour: number; minute: number }> {
  const normalized = timePoints.map(parseTimePoint);
  const unique = new Map(normalized.map((timePoint) => [formatTimePoint(timePoint.hour, timePoint.minute), timePoint]));

  return [...unique.values()].sort((a, b) => a.hour * 60 + a.minute - (b.hour * 60 + b.minute));
}

function parseTimePoint(timePoint: string): { hour: number; minute: number } {
  const matches = timePoint.match(/^([01]\d|2[0-3]):([0-5]\d)$/);
  if (!matches) {
    throw new Error("Invalid time point");
  }

  return {
    hour: parseInt(matches[1], 10),
    minute: parseInt(matches[2], 10),
  };
}

function formatTimePoint(hour: number, minute: number): string {
  return `${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")}`;
}
