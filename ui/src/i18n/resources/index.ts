import { type Resource } from "i18next";

import en from "./en";
import zh from "./zh";

export const LANG_EN = "en" as const;
export const LANG_ZH = "zh" as const;

const resources: Resource = {
  [LANG_EN]: {
    name: "English",
    translation: en,
  },
  [LANG_ZH]: {
    name: "简体中文",
    translation: zh,
  },
};

export default resources;
