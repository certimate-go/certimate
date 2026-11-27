import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, Form, Radio, type RadioChangeEvent, Select, theme } from "antd";

import { useAppLocaleMenuItems } from "@/components/AppLocale";
import { useAppThemeMenuItems } from "@/components/AppTheme";
import { useAntdForm, useAppSettings, useBrowserTheme } from "@/hooks";

import { withBasePath } from "@/utils/url";

const SettingsAppearance = () => {
  return (
    <>
      <div>
        <SettingsAppearanceTheme />
      </div>

      <div>
        <SettingsAppearanceLanguage />
      </div>

      <div>
        <SettingsAppearancePreferences />
      </div>
    </>
  );
};

const SettingsAppearanceTheme = () => {
  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { themeMode, setThemeMode } = useBrowserTheme();
  const themeItems = useAppThemeMenuItems();

  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm({
    initialValues: {
      value: themeMode,
    },
    onSubmit: async (values) => {
      setThemeMode(values.value);
      window.location.reload();
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const handleRadioChange = (e: RadioChangeEvent) => {
    setFormChanged(e.target.value !== themeMode);
  };

  return (
    <Card
      title={
        <div className="mt-4 mb-0 overflow-hidden">
          <h2 className="truncate text-base font-semibold">{t("settings.appearance.theme.title")}</h2>
          <p className="text-sm font-normal whitespace-normal text-foreground/70">{t("settings.appearance.theme.subtitle")}</p>
        </div>
      }
    >
      <Form className="md:max-w-160" {...formProps} form={formInst} disabled={formPending} layout="vertical">
        <Form.Item name="value" help={formChanged ? t("settings.appearance.theme.form.value.help") : void 0}>
          <Radio.Group block onChange={handleRadioChange}>
            <div className="flex w-full items-center gap-4 max-md:flex-wrap">
              {themeItems.map((item) => (
                <div className="relative max-w-44 flex-1/3 max-md:flex-1/2 max-sm:flex-1" key={item.key}>
                  <div className="overflow-hidden rounded-lg border border-solid" style={{ borderColor: themeToken.colorBorder }}>
                    <img className="mb-2 w-full" src={withBasePath(`/imgs/themes/${item.key}.png`)} />
                    <div className="mb-2 px-2">
                      <Radio value={item.key}>{item.label}</Radio>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Radio.Group>
        </Form.Item>

        <Form.Item noStyle>
          <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
            {t("common.button.save")}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

const SettingsAppearanceLanguage = () => {
  const { i18n, t } = useTranslation();

  const localeItems = useAppLocaleMenuItems();

  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm({
    initialValues: {
      value: i18n.resolvedLanguage ?? i18n.language,
    },
    onSubmit: async (values) => {
      i18n.changeLanguage(values.value);
      window.location.reload();
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const handleSelectChange = (value: string) => {
    setFormChanged(value !== (i18n.resolvedLanguage ?? i18n.language));
  };

  return (
    <Card
      title={
        <div className="mt-4 mb-0 overflow-hidden">
          <h2 className="truncate text-base font-semibold">{t("settings.appearance.language.title")}</h2>
          <p className="text-sm font-normal whitespace-normal text-foreground/70">{t("settings.appearance.language.subtitle")}</p>
        </div>
      }
    >
      <Form className="md:max-w-160" {...formProps} form={formInst} disabled={formPending} layout="vertical">
        <Form.Item name="value" help={formChanged ? t("settings.appearance.language.form.value.help") : void 0}>
          <Select
            options={localeItems.map((item) => ({
              key: item.key,
              value: item.key,
              label: item.label,
            }))}
            value={i18n.resolvedLanguage ?? i18n.language}
            onChange={handleSelectChange}
          />
        </Form.Item>

        <Form.Item noStyle>
          <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
            {t("common.button.save")}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

const SettingsAppearancePreferences = () => {
  const { t } = useTranslation();

  const { appSettings: globalAppSettings, setAppSettings: setGlobalAppSettings } = useAppSettings();

  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm({
    initialValues: {
      defaultPerPage: globalAppSettings.defaultPerPage,
      defaultWorkflowLayout: globalAppSettings.defaultWorkflowLayout,
    },
    onSubmit: async (values) => {
      setGlobalAppSettings({
        ...globalAppSettings,
        defaultPerPage: values.defaultPerPage,
        defaultWorkflowLayout: values.defaultWorkflowLayout,
      });
      setFormChanged(false);
    },
  });
  const [formChanged, setFormChanged] = useState(false);

  const handleFormChange = () => {
    const values = formInst.getFieldsValue();
    const changed = values.defaultPerPage !== globalAppSettings.defaultPerPage || values.defaultWorkflowLayout !== globalAppSettings.defaultWorkflowLayout;
    setFormChanged(changed);
  };

  return (
    <Card
      title={
        <div className="mt-4 mb-0 overflow-hidden">
          <h2 className="truncate text-base font-semibold">{t("settings.appearance.preferences.title")}</h2>
          <p className="text-sm font-normal whitespace-normal text-foreground/70">{t("settings.appearance.preferences.subtitle")}</p>
        </div>
      }
    >
      <Form className="md:max-w-160" {...formProps} form={formInst} disabled={formPending} layout="vertical" onValuesChange={handleFormChange}>
        <Form.Item name="defaultPerPage" label={t("settings.appearance.preferences.form.default_per_page.label")}>
          <Select
            options={[10, 15, 20, 30, 50, 100].map((value) => ({
              value: value,
              label: `${value} ${t("settings.appearance.preferences.form.default_per_page.unit")}`,
            }))}
            placeholder={t("settings.appearance.preferences.form.default_per_page.placeholder")}
          />
        </Form.Item>

        <Form.Item name="defaultWorkflowLayout" label={t("settings.appearance.preferences.form.default_workflow_layout.label")}>
          <Select
            options={["horizontal", "vertical"].map((value) => ({
              value: value,
              label: t(`settings.appearance.preferences.form.default_workflow_layout.option.${value}`),
            }))}
            placeholder={t("settings.appearance.preferences.form.default_workflow_layout.placeholder")}
          />
        </Form.Item>

        <Form.Item noStyle>
          <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
            {t("common.button.save")}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default SettingsAppearance;
