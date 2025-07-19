import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Divider, Form, Radio, type RadioChangeEvent, Select, theme } from "antd";

import { useAppLocaleMenuItems } from "@/components/AppLocale";
import { useAppThemeMenuItems } from "@/components/AppTheme";
import { useBrowserTheme } from "@/hooks";

const SettingsAppearance = () => {
  const { t } = useTranslation();

  return (
    <>
      <h2>{t("settings.appearance.theme.title")}</h2>
      <div className="md:max-w-160">
        <SettingsAppearanceTheme />
      </div>

      <Divider />

      <h2>{t("settings.appearance.language.title")}</h2>
      <div className="md:max-w-160">
        <SettingsAppearanceLanguage />
      </div>
    </>
  );
};

const SettingsAppearanceTheme = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { themeMode, setThemeMode } = useBrowserTheme();
  const themeItems = useAppThemeMenuItems();
  const [themeChanged, setThemeChanged] = useState(false);

  const handleThemeChange = (e: RadioChangeEvent) => {
    if (e.target.value !== themeMode) {
      setThemeChanged(true);
      setThemeMode(e.target.value);
    }
  };

  return (
    <div className={className} style={style}>
      <Form layout="vertical">
        <Form.Item extra={themeChanged ? t("settings.appearance.theme.form.value.extra") : undefined}>
          <Radio.Group className="w-full" value={themeMode} onChange={handleThemeChange}>
            <div className="flex w-full items-center gap-4 max-md:flex-wrap">
              {themeItems.map((item) => (
                <div className="relative max-w-44 flex-1/3 max-md:flex-1/2 max-sm:flex-1" key={item.key}>
                  <div className="overflow-hidden rounded-lg border border-solid" style={{ borderColor: themeToken.colorBorderSecondary }}>
                    <img className="mb-2 w-full" src={`/imgs/themes/${item.key}.png`} />
                    <div className="mb-2 px-2">
                      <Radio value={item.key}>{item.label}</Radio>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Radio.Group>
        </Form.Item>
      </Form>
    </div>
  );
};

const SettingsAppearanceLanguage = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { i18n, t } = useTranslation();

  const localeItems = useAppLocaleMenuItems();
  const [localeChanged, setLocaleChanged] = useState(false);

  const handleLocaleChange = (value: string) => {
    if (value !== i18n.language) {
      setLocaleChanged(true);
      i18n.changeLanguage(value);
    }
  };

  return (
    <div className={className} style={style}>
      <Form layout="vertical">
        <Form.Item extra={localeChanged ? t("settings.appearance.language.form.value.extra") : undefined}>
          <Select
            options={localeItems.map((item) => ({
              key: item.key,
              value: item.key,
              label: item.label,
            }))}
            value={i18n.language}
            onChange={handleLocaleChange}
          />
        </Form.Item>
      </Form>
    </div>
  );
};

export default SettingsAppearance;
