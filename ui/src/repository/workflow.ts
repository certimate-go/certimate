import { type RecordSubscription } from "pocketbase";

import { type WorkflowModel } from "@/domain/workflow";
import { COLLECTION_NAME_WORKFLOW, getPocketBase } from "./_pocketbase";

export type ListRequest = {
  keyword?: string;
  enabled?: boolean;
  sort?: string;
  page?: number;
  perPage?: number;
  expand?: boolean;
};

export const list = async (request: ListRequest) => {
  const pb = getPocketBase();

  const filters: string[] = [];
  if (request.keyword) {
    filters.push(pb.filter("name~{:keyword}", { keyword: request.keyword }));
  }
  if (request.enabled != null) {
    filters.push(pb.filter("enabled={:enabled}", { enabled: request.enabled }));
  }

  const sort = request.sort || "-created";

  const page = request.page || 1;
  const perPage = request.perPage || 10;

  return await pb.collection(COLLECTION_NAME_WORKFLOW).getList<WorkflowModel>(page, perPage, {
    expand: request.expand ? "lastRunRef" : undefined,
    filter: filters.join(" && "),
    sort: sort,
    requestKey: null,
  });
};

export const get = async (id: string) => {
  return await getPocketBase().collection(COLLECTION_NAME_WORKFLOW).getOne<WorkflowModel>(id, {
    expand: "lastRunRef",
    requestKey: null,
  });
};

export const save = async (record: MaybeModelRecord<WorkflowModel>) => {
  if (record.id) {
    return await getPocketBase()
      .collection(COLLECTION_NAME_WORKFLOW)
      .update<WorkflowModel>(record.id as string, record);
  }

  return await getPocketBase().collection(COLLECTION_NAME_WORKFLOW).create<WorkflowModel>(record);
};

export const remove = async (record: MaybeModelRecordWithId<WorkflowModel> | MaybeModelRecordWithId<WorkflowModel>[]) => {
  const pb = getPocketBase();

  if (Array.isArray(record)) {
    const batch = pb.createBatch();
    for (const item of record) {
      batch.collection(COLLECTION_NAME_WORKFLOW).delete(item.id);
    }
    const res = await batch.send();
    return res.every((e) => e.status === 200);
  } else {
    return await pb.collection(COLLECTION_NAME_WORKFLOW).delete(record.id);
  }
};

export const subscribe = async (id: string, cb: (e: RecordSubscription<WorkflowModel>) => void) => {
  return getPocketBase().collection(COLLECTION_NAME_WORKFLOW).subscribe(id, cb);
};

export const unsubscribe = async (id: string) => {
  return getPocketBase().collection(COLLECTION_NAME_WORKFLOW).unsubscribe(id);
};
