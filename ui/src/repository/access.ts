import dayjs from "dayjs";

import { type AccessModel } from "@/domain/access";
import { COLLECTION_NAME_ACCESS, getPocketBase } from "./_pocketbase";

const pb = getPocketBase();
const pbco = pb.collection(COLLECTION_NAME_ACCESS);

const _commonFields = ["id", "name", "provider", "reserve", "created", "updated", "deleted"];

export const list = async () => {
  const list = await pbco.getFullList<AccessModel>({
    batch: 65535,
    fields: [..._commonFields].join(","),
    filter: "deleted=null",
    sort: "-created",
    requestKey: null,
  });
  return {
    totalItems: list.length,
    items: list,
  };
};

export const get = async (id: string) => {
  return await pbco.getOne<AccessModel>(id, {
    requestKey: null,
  });
};

export const save = async (record: MaybeModelRecord<AccessModel>) => {
  if (record.id) {
    return await pbco.update<AccessModel>(record.id, record);
  }

  return await pbco.create<AccessModel>(record);
};

export const remove = async (record: MaybeModelRecordWithId<AccessModel> | MaybeModelRecordWithId<AccessModel>[]) => {
  const deletedAt = dayjs.utc().format("YYYY-MM-DD HH:mm:ss");

  if (Array.isArray(record)) {
    const batch = pb.createBatch();
    for (const item of record) {
      batch.collection(COLLECTION_NAME_ACCESS).update(item.id, { deleted: deletedAt });
    }
    const res = await batch.send();
    return res.every((e) => e.status >= 200 && e.status < 400);
  } else {
    await pbco.update<AccessModel>(record.id!, { deleted: deletedAt });
    return true;
  }
};
