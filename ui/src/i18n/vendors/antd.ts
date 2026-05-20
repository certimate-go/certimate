import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { type Locale } from "antd/es/locale";
import AntdLocaleEnUS from "antd/locale/en_US";
import AntdLocaleZhCN from "antd/locale/zh_CN";

import { localeNames } from "../locales";

const localesMap: Record<string, Locale> = {
  [localeNames.EN]: AntdLocaleEnUS,
  [localeNames.ZH]: AntdLocaleZhCN,
};

export const useAntdLocale = () => {
  const { i18n } = useTranslation();

  const [antdLocale, setAntdLocale] = useState<Locale>(localesMap[i18n.language]);

  useEffect(() => {
    setAntdLocale(localesMap[i18n.language]);
  }, [i18n.language]);

  return antdLocale;
};
