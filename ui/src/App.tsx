import "reflect-metadata";
import { useEffect, useState } from "react";
import { RouterProvider } from "react-router-dom";
import {
  IconCheck,
  IconChevronDown,
  IconChevronRight,
  IconCircleCheckFilled,
  IconCircleXFilled,
  IconExclamationCircleFilled,
  IconEye,
  IconEyeOff,
  IconInfoCircleFilled,
  IconSearch,
  IconX,
} from "@tabler/icons-react";
import { App, ConfigProvider, Empty, type ThemeConfig, theme } from "antd";

import { useBrowserTheme } from "@/hooks";
import { useAntdLocale, useDayjsLocale, useZodLocale } from "@/i18n";
import { router } from "@/routers";

const antdThemesMap: Record<string, ThemeConfig> = {
  ["light"]: { algorithm: theme.defaultAlgorithm },
  ["dark"]: { algorithm: theme.darkAlgorithm },
};

const RootApp = () => {
  const { theme: browserTheme } = useBrowserTheme();

  const [antdTheme, setAntdTheme] = useState(antdThemesMap[browserTheme]);
  const antdLocale = useAntdLocale();
  useDayjsLocale();
  useZodLocale();

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
      alert={{
        closeIcon: <IconX size="1.25em" />,
        errorIcon: <IconCircleXFilled size="1em" />,
        infoIcon: <IconInfoCircleFilled size="1em" />,
        successIcon: <IconCircleCheckFilled size="1em" />,
        warningIcon: <IconExclamationCircleFilled size="1em" />,
      }}
      collapse={{
        expandIcon: (panelProps) => {
          if (panelProps.showArrow != void 0 && !panelProps.showArrow) return void 0;
          return panelProps.isActive ? <IconChevronDown size="1.25em" /> : <IconChevronRight size="1.25em" />;
        },
      }}
      drawer={{
        closeIcon: <IconX size="1.25em" />,
      }}
      empty={{
        image: Empty.PRESENTED_IMAGE_SIMPLE,
      }}
      inputPassword={{
        iconRender: (visible) => {
          return visible ? <IconEye size="1.25em" /> : <IconEyeOff size="1.25em" />;
        },
      }}
      inputSearch={{
        searchIcon: <IconSearch size="1.25em" />,
      }}
      modal={{
        closeIcon: <IconX size="1.25em" />,
      }}
      notification={{
        closeIcon: <IconX size="1.25em" />,
      }}
      select={{
        menuItemSelectedIcon: <IconCheck size="1.25em" />,
        removeIcon: <IconX size="1.25em" />,
      }}
    >
      <App>
        <RouterProvider router={router} />
      </App>
    </ConfigProvider>
  );
};

export default RootApp;
