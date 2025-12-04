import { produce } from "immer";
import { create } from "zustand";

import { type NotifyTemplateContent, SETTINGS_NAMES, type ScriptTemplateContent, type SettingsModel } from "@/domain/settings";
import { get as getSettings, save as saveSettings } from "@/repository/settings";

import { type NotifyTemplatesState, type NotifyTemplatesStore, type ScriptTemplatesState, type ScriptTemplatesStore } from "./types";

export const useNotifyTemplatesStore = create<NotifyTemplatesStore>((set, get) => {
  let fetcher: Promise<SettingsModel<NotifyTemplateContent>> | null = null; // 防止多次重复请求
  let model: SettingsModel<NotifyTemplateContent>; // 记录当前设置的其他字段，保存回数据库时用

  return {
    templates: [],
    loading: false,
    loadedAtOnce: false,

    fetchTemplates: async (refresh = true) => {
      if (!refresh) {
        if (get().loadedAtOnce) {
          return;
        }
      }

      fetcher ??= getSettings(SETTINGS_NAMES.NOTIFY_TEMPLATE);

      try {
        set({ loading: true });
        model = await fetcher;
        set({ templates: model.content.templates ?? [], loadedAtOnce: true });
      } finally {
        fetcher = null;
        set({ loading: false });
      }
    },

    setTemplates: async (templates) => {
      model ??= await getSettings(SETTINGS_NAMES.NOTIFY_TEMPLATE);
      model = await saveSettings<NotifyTemplateContent>({
        ...model,
        content: {
          ...model.content,
          templates: templates,
        },
      });

      set(
        produce((state: NotifyTemplatesState) => {
          state.templates = model.content.templates ?? [];
          state.loadedAtOnce = true;
        })
      );
    },

    addTemplate: async (template) => {
      const templates = produce(get().templates, (draft) => {
        const index = draft.findIndex((t) => t.name === template.name);
        if (index !== -1) {
          draft[index] = template;
        } else {
          draft.push(template);
        }

        return draft;
      });
      get().setTemplates(templates);
    },

    removeTemplateByIndex: async (index) => {
      const templates = produce(get().templates, (draft) => {
        draft = draft.filter((_, i) => i !== index);
        return draft;
      });
      get().setTemplates(templates);
    },

    removeTemplateByName: async (name) => {
      const templates = produce(get().templates, (draft) => {
        draft = draft.filter((e) => e.name !== name);
        return draft;
      });
      get().setTemplates(templates);
    },
  };
});

export const useScriptTemplatesStore = create<ScriptTemplatesStore>((set, get) => {
  let fetcher: Promise<SettingsModel<ScriptTemplateContent>> | null = null; // 防止多次重复请求
  let model: SettingsModel<ScriptTemplateContent>; // 记录当前设置的其他字段，保存回数据库时用

  return {
    templates: [],
    loading: false,
    loadedAtOnce: false,

    fetchTemplates: async (refresh = true) => {
      if (!refresh) {
        if (get().loadedAtOnce) {
          return;
        }
      }

      fetcher ??= getSettings(SETTINGS_NAMES.SCRIPT_TEMPLATE);

      try {
        set({ loading: true });
        model = await fetcher;
        set({ templates: model.content.templates ?? [], loadedAtOnce: true });
      } finally {
        fetcher = null;
        set({ loading: false });
      }
    },

    setTemplates: async (templates) => {
      model ??= await getSettings(SETTINGS_NAMES.SCRIPT_TEMPLATE);
      model = await saveSettings<ScriptTemplateContent>({
        ...model,
        content: {
          ...model.content,
          templates: templates,
        },
      });

      set(
        produce((state: ScriptTemplatesState) => {
          state.templates = model.content.templates ?? [];
          state.loadedAtOnce = true;
        })
      );
    },

    addTemplate: async (template) => {
      const templates = produce(get().templates, (draft) => {
        const index = draft.findIndex((t) => t.name === template.name);
        if (index !== -1) {
          draft[index] = template;
        } else {
          draft.push(template);
        }

        return draft;
      });
      get().setTemplates(templates);
    },

    removeTemplateByIndex: async (index) => {
      const templates = produce(get().templates, (draft) => {
        draft = draft.filter((_, i) => i !== index);
        return draft;
      });
      get().setTemplates(templates);
    },

    removeTemplateByName: async (name) => {
      const templates = produce(get().templates, (draft) => {
        draft = draft.filter((e) => e.name !== name);
        return draft;
      });
      get().setTemplates(templates);
    },
  };
});
