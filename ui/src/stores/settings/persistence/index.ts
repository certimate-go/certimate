import { produce } from "immer";
import { create } from "zustand";

import { type PersistenceSettingsContent, SETTINGS_NAMES, type SettingsModel } from "@/domain/settings";
import { get as getSettings, save as saveSettings } from "@/repository/settings";

import { type PersistenceSettingsState, type PersistenceSettingsStore } from "./types";

export const usePersistenceSettingsStore = create<PersistenceSettingsStore>((set, get) => {
  let fetcher: Promise<SettingsModel<PersistenceSettingsContent>> | null = null; // 防止多次重复请求
  let model: SettingsModel<PersistenceSettingsContent>; // 记录当前设置的其他字段，保存回数据库时用

  return {
    settings: {} as PersistenceSettingsContent,
    loading: false,
    loadedAtOnce: false,

    loadSettings: async (refresh = true) => {
      if (!refresh) {
        if (get().loadedAtOnce) {
          return;
        }
      }

      fetcher ??= getSettings(SETTINGS_NAMES.PERSISTENCE);

      try {
        set({ loading: true });
        model = await fetcher;
        set({ settings: model.content, loadedAtOnce: true });
      } finally {
        fetcher = null;
        set({ loading: false });
      }
    },

    saveSettings: async (settings) => {
      model ??= await getSettings(SETTINGS_NAMES.PERSISTENCE);
      model = await saveSettings<PersistenceSettingsContent>({
        ...model,
        content: settings,
      });

      set(
        produce((state: PersistenceSettingsState) => {
          state.settings = model.content;
          state.loadedAtOnce = true;
        })
      );
    },
  };
});
