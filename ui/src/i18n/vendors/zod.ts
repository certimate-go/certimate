import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { config as zconfig } from "zod";
import { en as ZodLocaleEnUs, zhCN as ZodLocaleZhCN } from "zod/locales";

import { localeNames } from "../locales";

type Locale = ReturnType<typeof ZodLocaleEnUs>;

const localesMap: Record<string, Locale> = {
  [localeNames.EN]: {
    localeError: (() => {
      // REF: https://github.com/colinhacks/zod/blob/main/packages/zod/src/v4/locales/en.ts
      const l = ZodLocaleEnUs();

      return (issue) => {
        let errmsg = l.localeError(issue);
        if (typeof errmsg !== "string") {
          errmsg = errmsg?.message ?? "";
        }

        switch (issue.code) {
          case "too_big":
            {
              if (issue.origin === "string") {
                const unit = issue.maximum === 0 || issue.maximum === 1 ? "character" : "characters";
                return `Invalid string: expected ${issue.inclusive ? "at most" : "less than"} ${issue.maximum.toString()} ${unit}`;
              }
            }
            break;

          case "too_small":
            {
              if (issue.origin === "string") {
                const unit = issue.minimum === 0 || issue.minimum === 1 ? "character" : "characters";
                return `Invalid string: expected ${issue.inclusive ? "at least" : "more than"} ${issue.minimum.toString()} ${unit}`;
              }
            }
            break;
        }

        return errmsg.startsWith("invalid") ? errmsg : `Invalid value: ${errmsg.replace(/^./, (c) => c.toLowerCase())}`;
      };
    })(),
  },

  [localeNames.ZH]: {
    localeError: (() => {
      // REF: https://github.com/colinhacks/zod/blob/main/packages/zod/src/v4/locales/zh-CN.ts
      const l = ZodLocaleZhCN();

      return (issue) => {
        let errmsg = l.localeError(issue);
        if (typeof errmsg !== "string") {
          errmsg = errmsg?.message ?? "";
        }

        switch (issue.code) {
          case "too_big":
            {
              if (issue.origin === "string") {
                return `无效字符串：期望${issue.inclusive ? "至多" : "小于"} ${issue.maximum.toString()} 个字符`;
              }
            }
            break;

          case "too_small":
            {
              if (issue.origin === "string") {
                return `无效字符串：期望${issue.inclusive ? "至少" : "大于"} ${issue.minimum.toString()} 个字符`;
              }
            }
            break;

          case "invalid_format":
            {
              if (errmsg.startsWith("无效")) {
                return errmsg.replace(/^(无效)([0-9A-Za-z])/, "$1 $2");
              }
            }
            break;
        }

        return errmsg.startsWith("无效") ? errmsg : `无效输入：${errmsg}`;
      };
    })(),
  },
};

export const useZodLocale = () => {
  const { i18n } = useTranslation();
  const zodLocale = localesMap[i18n.resolvedLanguage ?? i18n.language];

  useEffect(() => {
    zconfig(zodLocale);
  }, [zodLocale]);

  return zodLocale;
};
