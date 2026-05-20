import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import dayjs from "dayjs";
import { type Locale as ILocale } from "dayjs/locale/en";
import "dayjs/locale/zh-cn";

import { localeNames } from "../locales";

type Locale = string | ILocale;

const localesMap: Record<string, Locale> = {
  [localeNames.EN]: "en",
  [localeNames.ZH]: "zh-cn",
};

export const useDayjsLocale = () => {
  const { i18n } = useTranslation();

  const [dayjsLocale, setDayjsLocale] = useState<Locale>(localesMap[i18n.language]);

  useEffect(() => {
    setDayjsLocale(localesMap[i18n.language]);
  }, [i18n.language]);

  useEffect(() => {
    dayjs.locale(dayjsLocale);
  }, [dayjsLocale]);

  return dayjsLocale;
};
