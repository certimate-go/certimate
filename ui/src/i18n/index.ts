import { initReactI18next } from "react-i18next";
import i18n from "i18next";
import i18nBrowserLanguageDetector from "i18next-browser-languagedetector";

import { localeNames } from "./locales";
import resources from "./resources";

const normalizeLanguage = (lng: string) => {
  const normalized = lng.toLowerCase().replace("_", "-");
  const base = normalized.split("-")[0] || localeNames.EN;

  if (base === localeNames.ZH) return localeNames.ZH;
  if (base === localeNames.EN) return localeNames.EN;
  return localeNames.EN;
};

i18n
  .use(i18nBrowserLanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    supportedLngs: [localeNames.ZH, localeNames.EN],
    fallbackLng: localeNames.EN,
    debug: true,
    interpolation: {
      escapeValue: false,
    },
    detection: {
      lookupLocalStorage: "certimate-ui-lang",
      convertDetectedLanguage: normalizeLanguage,
    },
  });

export { localeNames };
export const localeResources = resources;

export { useAntdLocale } from "./vendors/antd";
export { useDayjsLocale } from "./vendors/dayjs";
export { useZodLocale } from "./vendors/zod";

export default i18n;
