import { useTranslation } from "react-i18next";
import { type Locale } from "antd/es/locale";
import AntdLocaleEnUS from "antd/locale/en_US";
import AntdLocaleRuRU from "antd/locale/ru_RU";
import AntdLocaleZhCN from "antd/locale/zh_CN";

import { localeNames } from "../locales";

const localesMap: Record<string, Locale> = {
  [localeNames.EN]: AntdLocaleEnUS,
  [localeNames.ZH]: AntdLocaleZhCN,
  [localeNames.RU]: AntdLocaleRuRU,
};

export const useAntdLocale = () => {
  const { i18n } = useTranslation();
  return localesMap[i18n.resolvedLanguage ?? i18n.language];
};
