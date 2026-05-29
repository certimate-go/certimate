import { type WorkflowLogModel } from "@/domain/workflowLog";

import { COLLECTION_NAME_WORKFLOW_LOG, getPocketBase } from "./_pocketbase";

const pb = getPocketBase();
const pbco = pb.collection(COLLECTION_NAME_WORKFLOW_LOG);

export const listByWorkflowRunId = async (workflowRunId: string) => {
  const list = await pbco.getFullList<WorkflowLogModel>({
    batch: 65535,
    filter: pb.filter("runRef={:workflowRunId}", { workflowRunId }),
    sort: "timestamp",
    requestKey: null,
  });
  return {
    totalItems: list.length,
    items: list,
  };
};
