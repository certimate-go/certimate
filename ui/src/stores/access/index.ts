import { produce } from "immer";
import { create } from "zustand";

import { type AccessModel } from "@/domain/access";
import { list as listAccesses, remove as removeAccess, save as saveAccess } from "@/repository/access";

import { type AccessesState, type AccessesStore } from "./types";

export const useAccessesStore = create<AccessesStore>((set, get) => {
  let fetcher: Promise<AccessModel[]> | null = null; // 防止多次重复请求

  return {
    accesses: [],
    loading: false,
    loadedAtOnce: false,

    fetchAccesses: async (refresh = true) => {
      if (!refresh) {
        if (get().loadedAtOnce) {
          return get().accesses;
        }
      }

      fetcher ??= listAccesses().then((res) => res.items);

      try {
        set({ loading: true });
        const accesses = await fetcher;
        set({ accesses: accesses ?? [], loadedAtOnce: true });
      } finally {
        fetcher = null;
        set({ loading: false });
      }

      return get().accesses;
    },

    createAccess: async (access) => {
      const record = await saveAccess(access);
      set(
        produce((state: AccessesState) => {
          state.accesses.unshift(record);
        })
      );

      return record as AccessModel;
    },

    updateAccess: async (access) => {
      const record = await saveAccess(access);
      set(
        produce((state: AccessesState) => {
          const index = state.accesses.findIndex((e) => e.id === record.id);
          if (index !== -1) {
            state.accesses[index] = record;
          }
        })
      );

      return record as AccessModel;
    },

    deleteAccess: async (access) => {
      await removeAccess(access);
      if (Array.isArray(access)) {
        set(
          produce((state: AccessesState) => {
            state.accesses = state.accesses.filter((e) => !access.some((item) => item.id === e.id));
          })
        );
      } else {
        set(
          produce((state: AccessesState) => {
            state.accesses = state.accesses.filter((e) => e.id !== access.id);
          })
        );
      }

      return access as AccessModel;
    },
  };
});
