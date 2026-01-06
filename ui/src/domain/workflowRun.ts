import { type WorkflowGraph, type WorkflowModel } from "./workflow";

export interface WorkflowRunModel extends BaseModel {
  workflowRef: string;
  status: string;
  trigger: string;
  startedAt: ISO8601String;
  endedAt: ISO8601String;
  graph?: WorkflowGraph;
  error?: string;
  outputs?: Array<{
    type: string;
    name: string;
    value: string;
    valueType: string;
  }>;
  expand?: {
    workflowRef?: Pick<WorkflowModel, "id" | "name" | "description">;
  };
}

export const WORKFLOW_RUN_STATUSES = Object.freeze({
  PENDING: "pending",
  PROCESSING: "processing",
  SUCCEEDED: "succeeded",
  FAILED: "failed",
  CANCELED: "canceled",
} as const);

export type WorkflorRunStatusType = (typeof WORKFLOW_RUN_STATUSES)[keyof typeof WORKFLOW_RUN_STATUSES];
