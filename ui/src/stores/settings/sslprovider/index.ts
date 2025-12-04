import { produce } from "immer";
import { create } from "zustand";

import { SETTINGS_NAMES, type SSLProviderSettingsContent, type SettingsModel } from "@/domain/settings";
import { get as getSettings, save as saveSettings } from "@/repository/settings";

import { type SSLProviderSettingsState, type SSLProviderSettingsStore } from "./types";

export const useSSLProviderSettingsStore = create<SSLProviderSettingsStore>((set, get) => {
  let fetcher: Promise<SettingsModel<SSLProviderSettingsContent>> | null = null; // 防止多次重复请求
  let model: SettingsModel<SSLProviderSettingsContent>; // 记录当前设置的其他字段，保存回数据库时用

  return {
    settings: {} as SSLProviderSettingsContent,
    loading: false,
    loadedAtOnce: false,

    loadSettings: async (refresh = true) => {
      if (!refresh) {
        if (get().loadedAtOnce) {
          return;
        }
      }

      fetcher ??= getSettings(SETTINGS_NAMES.SSL_PROVIDER);

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
      model ??= await getSettings(SETTINGS_NAMES.SSL_PROVIDER);
      model = await saveSettings<SSLProviderSettingsContent>({
        ...model,
        content: settings,
      });

      set(
        produce((state: SSLProviderSettingsState) => {
          state.settings = model.content;
          state.loadedAtOnce = true;
        })
      );
    },
  };
});
