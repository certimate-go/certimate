import "reflect-metadata";
import { useEffect, useLayoutEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { RouterProvider } from "react-router-dom";
import { Alert, App, ConfigProvider, type ThemeConfig, theme } from "antd";
import { type Locale } from "antd/es/locale";
import AntdLocaleEnUS from "antd/locale/en_US";
import AntdLocaleZhCN from "antd/locale/zh_CN";
import dayjs from "dayjs";
import { z } from "zod";
import { en as ZodLocaleEnUs, zhCN as ZodLocaleZhCN } from "zod/locales";
import "dayjs/locale/zh-cn";

import { useBrowserTheme } from "@/hooks";
import { localeNames } from "@/i18n";
import { router } from "@/routers";

const antdLocalesMap: Record<string, Locale> = {
  [localeNames.EN]: AntdLocaleEnUS,
  [localeNames.ZH]: AntdLocaleZhCN,
};

const antdThemesMap: Record<string, ThemeConfig> = {
  ["light"]: { algorithm: theme.defaultAlgorithm },
  ["dark"]: { algorithm: theme.darkAlgorithm },
};

const zodLocalesMap: Record<string, typeof ZodLocaleEnUs> = {
  [localeNames.EN]: ZodLocaleEnUs,
  [localeNames.ZH]: ZodLocaleZhCN,
};

const RootApp = () => {
  const { i18n } = useTranslation();

  const { theme: browserTheme } = useBrowserTheme();

  const [antdLocale, setAntdLocale] = useState(antdLocalesMap[i18n.language]);
  const [antdTheme, setAntdTheme] = useState(antdThemesMap[browserTheme]);

  const handleLanguageChanged = () => {
    setAntdLocale(antdLocalesMap[i18n.language]);
    dayjs.locale(i18n.language);
    z.config(zodLocalesMap[i18n.language]?.());
  };

  i18n.on("initialized", handleLanguageChanged);
  i18n.on("languageChanged", handleLanguageChanged);
  useLayoutEffect(() => {
    handleLanguageChanged();

    return () => {
      i18n.off("initialized", handleLanguageChanged);
      i18n.off("languageChanged", handleLanguageChanged);
    };
  }, [i18n]);

  useEffect(() => {
    setAntdTheme(antdThemesMap[browserTheme]);

    const root = window.document.documentElement;
    root.classList.remove("light", "dark");
    root.classList.add(browserTheme);
  }, [browserTheme]);

  return (
    <ConfigProvider
      locale={antdLocale}
      theme={{
        ...antdTheme,
        token: {
          /* @see global.css, YOU MUST MODIFY BOTH DEFINITIONS AT THE SAME TIME! */
          colorBgBase: browserTheme === "dark" ? "#17191c" : "#ffffff",
          colorTextBase: browserTheme === "dark" ? "#fafaf9" : "#141414",
          colorPrimary: browserTheme === "dark" ? "#f97316" : "#ea580c",
          colorLink: browserTheme === "dark" ? "#f97316" : "#ea580c",
          colorInfo: browserTheme === "dark" ? "#478be6" : "#0969da",
          colorSuccess: browserTheme === "dark" ? "#57ab5a" : "#1a7f37",
          colorWarning: browserTheme === "dark" ? "#daaa3f" : "#eac54f",
          colorError: browserTheme === "dark" ? "#e5534b" : "#d1242f",

          /* @see https://tailwindcss.com/docs/responsive-design#overview */
          screenXS: 30 * 16,
          screenXSMin: 30 * 16,
          screenXSMax: 40 * 16 - 1,
          screenSM: 40 * 16,
          screenSMMin: 40 * 16,
          screenSMMax: 48 * 16 - 1,
          screenMD: 48 * 16,
          screenMDMin: 48 * 16,
          screenMDMax: 64 * 16 - 1,
          screenLG: 64 * 16,
          screenLGMin: 64 * 16,
          screenLGMax: 80 * 16 - 1,
          screenXL: 80 * 16,
          screenXLMin: 80 * 16,
          screenXLMax: 96 * 16 - 1,
          screenXXL: 96 * 16,
          screenXXLMin: 96 * 16,
          padding: 16,
          paddingXS: 8,
          paddingXXS: 6,
        },
        components: {
          Layout: {
            ...antdTheme?.components?.Layout,
            bodyBg: "transparent",
            headerBg: "transparent",
            siderBg: "transparent",
          },
          Dropdown: {
            ...antdTheme?.components?.Dropdown,
            paddingBlock: 9,
          },
          Form: {
            ...antdTheme?.components?.Form,
            itemMarginBottom: 28,
          },
        },
      }}
    >
      <Alert.ErrorBoundary>
        <App>
          <RouterProvider router={router} />
        </App>
      </Alert.ErrorBoundary>
    </ConfigProvider>
  );
};

export default RootApp;
