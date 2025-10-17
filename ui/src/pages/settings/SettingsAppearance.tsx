import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Divider, Form, Radio, type RadioChangeEvent, Select, theme } from "antd";
import { produce } from "immer";

import { useAppLocaleMenuItems } from "@/components/AppLocale";
import { useAppThemeMenuItems } from "@/components/AppTheme";
import { useAppSettings, useBrowserTheme } from "@/hooks";

const SettingsAppearance = () => {
  const { t } = useTranslation();

  return (
    <>
      <h2>{t("settings.appearance.theme.title")}</h2>
      <SettingsAppearanceTheme className="md:max-w-160" />

      <Divider />

      <h2>{t("settings.appearance.language.title")}</h2>
      <SettingsAppearanceLanguage className="md:max-w-160" />

      <Divider />

      <h2>{t("settings.appearance.pagination.title")}</h2>
      <SettingsAppearancePagination className="md:max-w-160" />

      <Divider />

      <h2>{t("settings.appearance.workflow.title")}</h2>
      <SettingsAppearanceWorkflow className="md:max-w-160" />
    </>
  );
};

const SettingsAppearanceTheme = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { themeMode, setThemeMode } = useBrowserTheme();
  const themeItems = useAppThemeMenuItems();
  const [themeChanged, setThemeChanged] = useState(false);

  const handleChange = (e: RadioChangeEvent) => {
    if (e.target.value !== themeMode) {
      setThemeChanged(true);
      setThemeMode(e.target.value);
    }
  };

  return (
    <div className={className} style={style}>
      <Form layout="vertical">
        <Form.Item extra={themeChanged ? t("settings.appearance.theme.form.value.extra") : void 0}>
          <Radio.Group block value={themeMode} onChange={handleChange}>
            <div className="flex w-full items-center gap-4 max-md:flex-wrap">
              {themeItems.map((item) => (
                <div className="relative max-w-44 flex-1/3 max-md:flex-1/2 max-sm:flex-1" key={item.key}>
                  <div className="overflow-hidden rounded-lg border border-solid" style={{ borderColor: themeToken.colorBorder }}>
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

  const handleChange = (value: string) => {
    if (value !== (i18n.resolvedLanguage ?? i18n.language)) {
      setLocaleChanged(true);
      i18n.changeLanguage(value);
    }
  };

  return (
    <div className={className} style={style}>
      <Form layout="vertical">
        <Form.Item extra={localeChanged ? t("settings.appearance.language.form.value.extra") : void 0}>
          <Select
            options={localeItems.map((item) => ({
              key: item.key,
              value: item.key,
              label: item.label,
            }))}
            value={i18n.resolvedLanguage ?? i18n.language}
            onChange={handleChange}
          />
        </Form.Item>
      </Form>
    </div>
  );
};

const SettingsAppearancePagination = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { appSettings: globalAppSettings, setAppSettings: setGlobalAppSettings } = useAppSettings();

  const handleChange = (value: (typeof globalAppSettings)["defaultPerPage"]) => {
    setGlobalAppSettings(
      produce(globalAppSettings, (draft) => {
        draft.defaultPerPage = value;
      })
    );
  };

  return (
    <div className={className} style={style}>
      <Form layout="vertical">
        <Form.Item label={t("settings.appearance.pagination.form.default_per_page.label")}>
          <Select
            options={[10, 15, 20, 30, 50, 100].map((value) => ({
              key: value,
              value: value,
              label: `${value} ${t("settings.appearance.pagination.form.default_per_page.unit")}`,
            }))}
            placeholder={t("settings.appearance.pagination.form.default_per_page.placeholder")}
            defaultValue={globalAppSettings.defaultPerPage}
            onChange={handleChange}
          />
        </Form.Item>
      </Form>
    </div>
  );
};

const SettingsAppearanceWorkflow = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const { t } = useTranslation();

  const { appSettings: globalAppSettings, setAppSettings: setGlobalAppSettings } = useAppSettings();

  const handleChange = (value: (typeof globalAppSettings)["defaultWorkflowLayout"]) => {
    setGlobalAppSettings(
      produce(globalAppSettings, (draft) => {
        draft.defaultWorkflowLayout = value;
      })
    );
  };

  return (
    <div className={className} style={style}>
      <Form layout="vertical">
        <Form.Item label={t("settings.appearance.workflow.form.default_designer_layout.label")}>
          <Select
            options={["horizontal", "vertical"].map((value) => ({
              key: value,
              value: value,
              label: t(`settings.appearance.workflow.form.default_designer_layout.option.${value}`),
            }))}
            placeholder={t("settings.appearance.workflow.form.default_designer_layout.placeholder")}
            defaultValue={globalAppSettings.defaultWorkflowLayout}
            onChange={handleChange}
          />
        </Form.Item>
      </Form>
    </div>
  );
};

export default SettingsAppearance;
