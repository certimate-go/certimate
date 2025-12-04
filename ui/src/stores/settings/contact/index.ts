import { produce } from "immer";
import { create } from "zustand";

import { type EmailsSettingsContent, SETTINGS_NAMES, type SettingsModel } from "@/domain/settings";
import { get as getSettings, save as saveSettings } from "@/repository/settings";

import { type ContactEmailsState, type ContactEmailsStore } from "./types";

export const useContactEmailsStore = create<ContactEmailsStore>((set, get) => {
  let fetcher: Promise<SettingsModel<EmailsSettingsContent>> | null = null; // 防止多次重复请求
  let model: SettingsModel<EmailsSettingsContent>; // 记录当前设置的其他字段，保存回数据库时用

  return {
    emails: [],
    loading: false,
    loadedAtOnce: false,

    fetchEmails: async (refresh = true) => {
      if (!refresh) {
        if (get().loadedAtOnce) {
          return get().emails;
        }
      }

      fetcher ??= getSettings(SETTINGS_NAMES.EMAILS);

      try {
        set({ loading: true });
        model = await fetcher;
        set({ emails: model.content.emails?.filter((s) => !!s)?.sort() ?? [], loadedAtOnce: true });
      } finally {
        fetcher = null;
        set({ loading: false });
      }

      return get().emails;
    },

    setEmails: async (emails) => {
      model ??= await getSettings(SETTINGS_NAMES.EMAILS);
      model = await saveSettings<EmailsSettingsContent>({
        ...model,
        content: {
          ...model.content,
          emails: emails,
        },
      });

      set(
        produce((state: ContactEmailsState) => {
          state.emails = model.content.emails?.sort() ?? [];
          state.loadedAtOnce = true;
        })
      );
    },

    addEmail: async (email) => {
      const emails = produce(get().emails, (draft) => {
        if (draft.includes(email)) return;
        draft.push(email);
        draft.sort();
        return draft;
      });
      get().setEmails(emails);
    },

    removeEmail: async (email) => {
      const emails = produce(get().emails, (draft) => {
        draft = draft.filter((e) => e !== email);
        draft.sort();
        return draft;
      });
      get().setEmails(emails);
    },
  };
});
