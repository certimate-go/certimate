import { initReactI18next } from "react-i18next";
import i18n from "i18next";
import i18nBrowserLanguageDetector from "i18next-browser-languagedetector";

import resources, { LOCALE_EN_NAME, LOCALE_ZH_NAME } from "./locales";

i18n
  .use(i18nBrowserLanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: LOCALE_ZH_NAME,
    debug: true,
    interpolation: {
      escapeValue: false,
    },
    detection: {
      lookupLocalStorage: "certimate-ui-lang",
    },
  });

export const localeNames = {
  ZH: LOCALE_ZH_NAME,
  EN: LOCALE_EN_NAME,
};

export const localeResources = resources;

export default i18n;
