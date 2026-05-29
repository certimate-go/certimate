import { type RecordSubscription } from "pocketbase";

import { type WorkflowModel } from "@/domain/workflow";
import { COLLECTION_NAME_WORKFLOW, getPocketBase } from "./_pocketbase";

const pb = getPocketBase();
const pbco = pb.collection(COLLECTION_NAME_WORKFLOW);

const _commonFields = [
  "id",
  "name",
  "description",
  "trigger",
  "triggerCron",
  "enabled",
  "hasDraft",
  "hasContent",
  "lastRunRef",
  "lastRunStatus",
  "lastRunTime",
  "created",
  "updated",
  "deleted",
];
const _expandFields = [
  "expand.lastRunRef.id",
  "expand.lastRunRef.status",
  "expand.lastRunRef.trigger",
  "expand.lastRunRef.startedAt",
  "expand.lastRunRef.endedAt",
  "expand.lastRunRef.error",
];

export const list = async ({
  keyword,
  enabled,
  sort = "-created",
  page = 1,
  perPage = 10,
  expand = false,
}: {
  keyword?: string;
  enabled?: boolean;
  sort?: string;
  page?: number;
  perPage?: number;
  expand?: boolean;
}) => {
  const filters: string[] = [];
  if (keyword) {
    filters.push(pb.filter("(id={:keyword} || name~{:keyword})", { keyword: keyword }));
  }
  if (enabled != null) {
    filters.push(pb.filter("enabled={:enabled}", { enabled: enabled }));
  }

  return await pbco.getList<WorkflowModel>(page, perPage, {
    expand: expand ? ["lastRunRef"].join(",") : void 0,
    fields: [..._commonFields, ..._expandFields].join(","),
    filter: filters.join(" && "),
    sort: sort || "-created",
    requestKey: null,
  });
};

export const get = async (id: string) => {
  return await pbco.getOne<WorkflowModel>(id, {
    expand: ["lastRunRef"].join(","),
    fields: ["*", ..._expandFields].join(","),
    requestKey: null,
  });
};

export const save = async (record: MaybeModelRecord<WorkflowModel>) => {
  if (record.id) {
    return await pbco.update<WorkflowModel>(record.id as string, record);
  }

  return await pbco.create<WorkflowModel>(record);
};

export const remove = async (record: MaybeModelRecordWithId<WorkflowModel> | MaybeModelRecordWithId<WorkflowModel>[]) => {
  if (Array.isArray(record)) {
    const batch = pb.createBatch();
    for (const item of record) {
      batch.collection(COLLECTION_NAME_WORKFLOW).delete(item.id);
    }
    const res = await batch.send();
    return res.every((e) => e.status >= 200 && e.status < 400);
  } else {
    return await pbco.delete(record.id);
  }
};

export const subscribe = async (id: string, cb: (e: RecordSubscription<WorkflowModel>) => void) => {
  return pbco.subscribe(id, cb);
};

export const unsubscribe = async (id: string) => {
  return pbco.unsubscribe(id);
};
