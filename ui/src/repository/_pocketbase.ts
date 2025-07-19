import PocketBase from "pocketbase";

let pb: PocketBase;
export const getPocketBase = () => {
  if (pb) return pb;
  pb = new PocketBase("/");
  pb.afterSend = (res, data) => {
    if (res.status === 401 && pb.authStore?.isValid) {
      pb.authStore.clear();
    }
    return data;
  };
  return pb;
};

export const COLLECTION_NAME_ADMIN = "_superusers";
export const COLLECTION_NAME_ACCESS = "access";
export const COLLECTION_NAME_CERTIFICATE = "certificate";
export const COLLECTION_NAME_SETTINGS = "settings";
export const COLLECTION_NAME_WORKFLOW = "workflow";
export const COLLECTION_NAME_WORKFLOW_RUN = "workflow_run";
export const COLLECTION_NAME_WORKFLOW_OUTPUT = "workflow_output";
export const COLLECTION_NAME_WORKFLOW_LOG = "workflow_logs";
