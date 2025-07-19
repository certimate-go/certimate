import { memo, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  IconBrandGithub,
  IconFingerprint,
  IconHelpCircle,
  IconHome,
  IconLogout,
  IconMenu2,
  IconSchema,
  IconSettings,
  IconShieldCheckered,
} from "@tabler/icons-react";
import { Alert, Button, Drawer, Layout, Menu, type MenuProps, theme } from "antd";

import AppLocale, { useAppLocaleMenuItems } from "@/components/AppLocale";
import AppTheme, { useAppThemeMenuItems } from "@/components/AppTheme";
import AppVersion from "@/components/AppVersion";
import Show from "@/components/Show";
import { APP_DOCUMENT_URL, APP_REPO_URL } from "@/domain/app";
import { useTriggerElement } from "@/hooks";
import { getAuthStore } from "@/repository/admin";
import { isBrowserHappy } from "@/utils/browser";

const ConsoleLayout = () => {
  const navigate = useNavigate();

  const { i18n, t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const localeMenuItems = useAppLocaleMenuItems();
  const themeMenuItems = useAppThemeMenuItems();

  const handleLogoutClick = () => {
    auth.clear();
    navigate("/login");
  };

  const handleDocumentClick = () => {
    if (i18n.language.startsWith("en")) {
      window.open(APP_DOCUMENT_URL + "/en/", "_blank");
    } else {
      window.open(APP_DOCUMENT_URL, "_blank");
    }
  };

  const handleGitHubClick = () => {
    window.open(APP_REPO_URL, "_blank");
  };

  const auth = getAuthStore();
  if (!auth.isValid || !auth.isSuperuser) {
    return <Navigate to="/login" />;
  }

  return (
    <Layout className="h-screen bg-background text-foreground">
      <Show when={!isBrowserHappy()}>
        <Alert message={t("common.text.happy_browser")} type="warning" showIcon closable />
      </Show>

      <Layout className="h-screen" hasSider>
        <Layout.Sider
          className="z-20 h-full border-r bg-background max-md:static max-md:hidden"
          style={{ borderColor: themeToken.colorBorderSecondary }}
          theme="light"
          width="256px"
        >
          <div className="flex size-full flex-col items-center justify-between overflow-hidden select-none">
            <div className="w-full px-2">
              <SiderMenu />
            </div>
            <div className="w-full px-2">
              <Menu
                style={{ background: "transparent", borderInlineEnd: "none" }}
                items={[
                  {
                    key: "theme",
                    icon: (
                      <span className="anticon scale-125">
                        <AppTheme.Icon size="1em" />
                      </span>
                    ),
                    label: t("common.menu.theme"),
                    children: themeMenuItems,
                  },
                  {
                    key: "locale",
                    icon: (
                      <span className="anticon scale-115">
                        <AppLocale.Icon size="1em" />
                      </span>
                    ),
                    label: t("common.menu.locale"),
                    children: localeMenuItems,
                  },
                  {
                    key: "document",
                    icon: (
                      <span className="anticon scale-125">
                        <IconHelpCircle size="1em" />
                      </span>
                    ),
                    label: t("common.menu.gethelp"),
                    onClick: handleDocumentClick,
                  },
                  {
                    key: "logout",
                    danger: true,
                    icon: (
                      <span className="anticon scale-125">
                        <IconLogout size="1em" />
                      </span>
                    ),
                    label: t("common.menu.logout"),
                    onClick: handleLogoutClick,
                  },
                ]}
                mode="vertical"
                selectable={false}
              />
            </div>
          </div>
        </Layout.Sider>

        <Layout className="flex flex-col overflow-hidden">
          <Layout.Header className="shadow-xs md:hidden" style={{ padding: 0 }}>
            <div className="flex size-full items-center justify-between overflow-hidden px-4">
              <div className="flex items-center gap-4">
                <SiderMenuDrawer trigger={<Button icon={<IconMenu2 size="1.25em" stroke="1.25" />} />} />
              </div>
              <div className="flex size-full grow items-center justify-end gap-4 overflow-hidden">
                <AppTheme.Dropdown>
                  <Button icon={<AppTheme.Icon size="1.25em" stroke="1.25" />} />
                </AppTheme.Dropdown>
                <AppLocale.Dropdown>
                  <Button icon={<AppLocale.Icon size="1.25em" stroke="1.25" />} />
                </AppLocale.Dropdown>
                <AppVersion.Badge>
                  <Button icon={<IconBrandGithub size="1.25em" stroke="1.25" />} onClick={handleGitHubClick} />
                </AppVersion.Badge>
                <Button danger icon={<IconLogout size="1.25em" stroke="1.25" />} onClick={handleLogoutClick} />
              </div>
            </div>
          </Layout.Header>

          <Layout.Content className="relative flex-1 overflow-x-hidden overflow-y-auto">
            <Outlet />
          </Layout.Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

const SiderMenu = memo(({ onSelect }: { onSelect?: (key: string) => void }) => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const MENU_KEY_HOME = "/";
  const MENU_KEY_WORKFLOWS = "/workflows";
  const MENU_KEY_CERTIFICATES = "/certificates";
  const MENU_KEY_ACCESSES = "/accesses";
  const MENU_KEY_SETTINGS = "/settings";
  const menuItems: Required<MenuProps>["items"] = [
    [MENU_KEY_HOME, <IconHome size="1em" />, t("dashboard.page.title")],
    [MENU_KEY_WORKFLOWS, <IconSchema size="1em" />, t("workflow.page.title")],
    [MENU_KEY_CERTIFICATES, <IconShieldCheckered size="1em" />, t("certificate.page.title")],
    [MENU_KEY_ACCESSES, <IconFingerprint size="1em" />, t("access.page.title")],
    [MENU_KEY_SETTINGS, <IconSettings size="1em" />, t("settings.page.title")],
  ].map(([key, icon, label]) => {
    return {
      key: key as string,
      icon: <span className="anticon scale-125">{icon}</span>,
      label: label,
      onClick: () => {
        navigate(key as string);
        onSelect?.(key as string);
      },
    };
  });
  const [menuSelectedKey, setMenuSelectedKey] = useState<string>();

  const getActiveMenuItem = () => {
    const item =
      menuItems.find((item) => item!.key === location.pathname) ??
      menuItems.find((item) => item!.key !== MENU_KEY_HOME && location.pathname.startsWith(item!.key as string));
    return item;
  };

  useEffect(() => {
    const item = getActiveMenuItem();
    if (item) {
      setMenuSelectedKey(item.key as string);
    } else {
      setMenuSelectedKey(undefined);
    }
  }, [location.pathname]);

  useEffect(() => {
    if (menuSelectedKey && menuSelectedKey !== getActiveMenuItem()?.key) {
      navigate(menuSelectedKey);
    }
  }, [menuSelectedKey]);

  return (
    <>
      <div className="flex w-full items-center gap-2 overflow-hidden px-4 py-2 max-md:py-0">
        <img src="/logo.svg" className="size-[36px]" />
        <span className="h-[64px] w-[81px] truncate text-base leading-[64px] font-semibold">Certimate</span>
        <AppVersion.LinkButton className="text-xs" />
      </div>
      <div className="w-full grow overflow-x-hidden overflow-y-auto">
        <Menu
          style={{ background: "transparent", borderInlineEnd: "none" }}
          items={menuItems}
          mode="vertical"
          selectedKeys={menuSelectedKey ? [menuSelectedKey] : []}
          onSelect={({ key }) => {
            setMenuSelectedKey(key);
          }}
        />
      </div>
    </>
  );
});

const SiderMenuDrawer = memo(({ trigger }: { trigger: React.ReactNode }) => {
  const { token: themeToken } = theme.useToken();

  const [siderOpen, setSiderOpen] = useState(false);

  const triggerEl = useTriggerElement(trigger, { onClick: () => setSiderOpen(true) });

  return (
    <>
      {triggerEl}

      <Drawer
        closable={false}
        destroyOnHidden
        open={siderOpen}
        placement="left"
        styles={{
          content: { paddingTop: themeToken.paddingSM, paddingBottom: themeToken.paddingSM },
          body: { padding: 0 },
        }}
        onClose={() => setSiderOpen(false)}
      >
        <SiderMenu onSelect={() => setSiderOpen(false)} />
      </Drawer>
    </>
  );
});

export default ConsoleLayout;
