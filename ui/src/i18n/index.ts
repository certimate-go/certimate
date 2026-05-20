import { initReactI18next } from "react-i18next";
import i18n from "i18next";
import i18nBrowserLanguageDetector from "i18next-browser-languagedetector";

import { localeNames } from "./locales";
import resources from "./resources";

i18n
  .use(i18nBrowserLanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: localeNames.EN,
    debug: true,
    interpolation: {
      escapeValue: false,
    },
    detection: {
      lookupLocalStorage: "certimate-ui-lang",
    },
  });

export { localeNames };
export const localeResources = resources;

export { useAntdLocale } from "./vendors/antd";
export { useDayjsLocale } from "./vendors/dayjs";
export { useZodLocale } from "./vendors/zod";

export default i18n;
