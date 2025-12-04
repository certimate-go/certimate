import { type AccessModel } from "@/domain/access";

export interface AccessesState {
  accesses: AccessModel[];
  loading: boolean;
  loadedAtOnce: boolean;
}

export interface AccessesActions {
  fetchAccesses: (refresh?: boolean) => Promise<AccessModel[]>;
  createAccess: (access: MaybeModelRecord<AccessModel>) => Promise<AccessModel>;
  updateAccess: (access: MaybeModelRecordWithId<AccessModel>) => Promise<AccessModel>;
  deleteAccess: (access: MaybeModelRecordWithId<AccessModel> | MaybeModelRecordWithId<AccessModel>[]) => Promise<AccessModel>;
}

export interface AccessesStore extends AccessesState, AccessesActions {}
