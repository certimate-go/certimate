import { WORKFLOW_TRIGGERS } from "@/domain/workflow";

import { get as httpGet, post as httpPost } from "./_api";

export const getStats = () => {
  type RespData = {
    concurrency: number;
    pendingRunIds: string[];
    processingRunIds: string[];
  };

  return httpGet<RespData>({
    url: `/api/workflows/stats`,
  });
};

export const startRun = (workflowId: string) => {
  return httpPost({
    url: `/api/workflows/${encodeURIComponent(workflowId)}/runs`,
    body: {
      trigger: WORKFLOW_TRIGGERS.MANUAL,
    },
  });
};

export const cancelRun = (workflowId: string, runId: string) => {
  return httpPost({
    url: `/api/workflows/${encodeURIComponent(workflowId)}/runs/${encodeURIComponent(runId)}/cancel`,
  });
};
