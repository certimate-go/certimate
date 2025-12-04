import { memo, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  IconBrandGithub,
  IconCertificate,
  IconCodeDots,
  IconFingerprint,
  IconHelpCircle,
  IconHierarchy3,
  IconHome,
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarRightCollapse,
  IconMenu2,
  IconPower,
  IconSettings,
} from "@tabler/icons-react";
import { Alert, Button, Drawer, Layout, Menu, type MenuProps, theme } from "antd";

import AppLocale from "@/components/AppLocale";
import AppTheme from "@/components/AppTheme";
import AppVersion from "@/components/AppVersion";
import Show from "@/components/Show";
import { APP_DOCUMENT_URL, APP_REPO_URL } from "@/domain/app";
import { useTriggerElement } from "@/hooks";
import { getAuthStore } from "@/repository/admin";
import { isBrowserHappy } from "@/utils/browser";

const ConsoleLayout = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const [siderCollapsed, setSiderCollapsed] = useState(false);

  const handleLogoutClick = () => {
    auth.clear();
    navigate("/login");
  };

  const handleDocumentClick = () => {
    window.open(APP_DOCUMENT_URL, "_blank");
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
        <Alert banner message={t("common.text.happy_browser")} type="warning" showIcon closable />
      </Show>

      <Layout className="h-screen" hasSider>
        <Layout.Sider
          className="group/sider z-20 h-full border-r bg-background max-md:static max-md:hidden"
          style={{ borderColor: themeToken.colorBorderSecondary }}
          theme="light"
          width={siderCollapsed ? 81 : 256}
        >
          <div className="flex size-full flex-col items-center justify-between overflow-hidden select-none">
            <div className="w-full px-2">
              <SiderMenu collapsed={siderCollapsed} />
            </div>
            <div className="w-full px-2 pb-2">
              <Menu
                style={{ background: "transparent", borderInlineEnd: "none" }}
                inlineCollapsed={siderCollapsed}
                items={[
                  {
                    key: "document",
                    icon: (
                      <span className="anticon scale-125" role="img">
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
                      <span className="anticon scale-125" role="img">
                        <IconPower size="1em" />
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
          <div className="absolute top-1/2 right-0 translate-x-1/2 -translate-y-1/2 opacity-0 transition-opacity group-hover/sider:opacity-100">
            <Button
              className="bg-background shadow"
              icon={
                siderCollapsed ? (
                  <IconLayoutSidebarRightCollapse size="1.5em" stroke="1.25" color="#999" />
                ) : (
                  <IconLayoutSidebarLeftCollapse size="1.5em" stroke="1.25" color="#999" />
                )
              }
              shape="circle"
              type="text"
              onClick={() => setSiderCollapsed(!siderCollapsed)}
            />
          </div>
        </Layout.Sider>

        <Layout className="flex flex-col overflow-hidden">
          <Layout.Header
            className="relative border-b shadow-sm md:hidden"
            style={{
              padding: 0,
              borderBottomColor: themeToken.colorBorderSecondary,
            }}
          >
            <div className="absolute inset-0 z-0">
              <div
                className="h-full w-full"
                style={{
                  backgroundImage:
                    "linear-gradient(rgba(255, 255, 255, 0.063) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.063) 1px, transparent 1px)",
                  backgroundSize: "20px 20px",
                }}
              >
                <div className="h-full w-full backdrop-blur-[1px]"></div>
              </div>
            </div>
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
                <Button danger icon={<IconPower size="1.25em" stroke="1.25" />} onClick={handleLogoutClick} />
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

const SiderMenu = memo(({ collapsed, onSelect }: { collapsed?: boolean; onSelect?: (key: string) => void }) => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const MENU_KEY_HOME = "/";
  const MENU_KEY_WORKFLOWS = "/workflows";
  const MENU_KEY_CERTIFICATES = "/certificates";
  const MENU_KEY_ACCESSES = "/accesses";
  const MENU_KEY_PRESETS = "/presets";
  const MENU_KEY_SETTINGS = "/settings";
  const menuItems: Required<MenuProps>["items"] = (
    [
      [MENU_KEY_HOME, "dashboard.page.title", <IconHome size="1em" />],
      [MENU_KEY_WORKFLOWS, "workflow.page.title", <IconHierarchy3 size="1em" />],
      [MENU_KEY_CERTIFICATES, "certificate.page.title", <IconCertificate size="1em" />],
      [MENU_KEY_ACCESSES, "access.page.title", <IconFingerprint size="1em" />],
      [MENU_KEY_PRESETS, "preset.page.title", <IconCodeDots size="1em" />],
      [MENU_KEY_SETTINGS, "settings.page.title", <IconSettings size="1em" />],
    ] satisfies Array<[string, string, React.ReactNode]>
  ).map(([key, label, icon]) => {
    return {
      key: key,
      icon: (
        <span className="anticon scale-125" role="img">
          {icon}
        </span>
      ),
      label: t(label),
      onClick: () => {
        navigate(key);
        onSelect?.(key);
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
      setMenuSelectedKey(void 0);
    }
  }, [location.pathname]);

  useEffect(() => {
    if (menuSelectedKey && menuSelectedKey !== getActiveMenuItem()?.key) {
      navigate(menuSelectedKey);
    }
  }, [menuSelectedKey]);

  return (
    <>
      <div className="h-[64px] w-full overflow-hidden px-4 py-2 max-md:py-0">
        <div className="flex size-full items-center justify-around gap-2">
          <img src="/logo.svg" className="size-[36px]" />
          <Show when={!collapsed}>
            <span className="w-[81px] truncate text-base leading-[64px] font-semibold">Certimate</span>
            <AppVersion.LinkButton className="text-xs" />
          </Show>
        </div>
      </div>
      <div className="w-full grow overflow-x-hidden overflow-y-auto">
        <Menu
          style={{ background: "transparent", borderInlineEnd: "none" }}
          inlineCollapsed={collapsed}
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

  const handleMenuSelect = useCallback(() => {
    setSiderOpen(false);
  }, []);

  return (
    <>
      {triggerEl}

      <Drawer
        closable={false}
        destroyOnHidden
        open={siderOpen}
        placement="left"
        styles={{
          section: { paddingTop: themeToken.paddingSM, paddingBottom: themeToken.paddingSM },
          body: { padding: 0 },
        }}
        onClose={() => setSiderOpen(false)}
      >
        <SiderMenu onSelect={handleMenuSelect} />
      </Drawer>
    </>
  );
});

export default ConsoleLayout;
