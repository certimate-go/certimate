import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { config as zconfig } from "zod";
import { en as ZodLocaleEnUs, zhCN as ZodLocaleZhCN } from "zod/locales";

import { localeNames } from "../locales";

type Locale = typeof ZodLocaleEnUs;

const localesMap: Record<string, Locale> = {
  [localeNames.EN]: ZodLocaleEnUs,
  [localeNames.ZH]: ZodLocaleZhCN,
};

export const useZodLocale = () => {
  const { i18n } = useTranslation();

  const [zodLocale, setZodLocale] = useState<Locale>(localesMap[i18n.language]);

  useEffect(() => {
    setZodLocale(localesMap[i18n.language]);
  }, [i18n.language]);

  useEffect(() => {
    if (typeof zodLocale === "function") {
      zconfig(zodLocale());
    }
  }, [zodLocale]);

  return zodLocale;
};
