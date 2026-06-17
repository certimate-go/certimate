import { type Resource } from "i18next";

import en from "./en";
import ru from "./ru";
import zh from "./zh";

export const LANG_ZH = "zh" as const;
export const LANG_EN = "en" as const;
export const LANG_RU = "ru" as const;

const resources: Resource = {
  [LANG_ZH]: {
    name: "简体中文",
    translation: zh,
  },
  [LANG_EN]: {
    name: "English",
    translation: en,
  },
  [LANG_RU]: {
    name: "Русский",
    translation: ru,
  },
};

export default resources;
