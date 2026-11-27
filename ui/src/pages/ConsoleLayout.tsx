import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  IconBrandGithub,
  IconCertificate,
  IconCodeDots,
  IconHelpCircle,
  IconHierarchy3,
  IconHome,
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarRightCollapse,
  IconMenu2,
  IconPlugConnected,
  IconPower,
  IconPuzzle,
  IconRobot,
  IconSettings,
} from "@tabler/icons-react";
import { Alert, Button, Drawer, Layout, Menu, type MenuProps, Tooltip, theme } from "antd";

import AppLocale from "@/components/AppLocale";
import AppTheme from "@/components/AppTheme";
import AppVersion from "@/components/AppVersion";
import Show from "@/components/Show";
import { APP_DOCUMENT_URL, APP_REPO_URL } from "@/domain/app";
import { useTriggerElement } from "@/hooks";
import { getAuthStore } from "@/repository/admin";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { isBrowserHappy } from "@/utils/browser";
import { mergeCls } from "@/utils/css";
import { withBasePath } from "@/utils/url";

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
        <Alert banner closable showIcon title={t("common.text.happy_browser")} type="warning" />
      </Show>

      <Layout className="h-screen" hasSider>
        <Layout.Sider
          className="z-20 h-full border-r bg-background max-md:static max-md:hidden"
          style={{ borderColor: themeToken.colorBorderSecondary }}
          theme="light"
          width={siderCollapsed ? 81 : 256}
        >
          <div className="flex size-full flex-col items-center justify-between overflow-hidden select-none">
            <div className="w-full px-2">
              <SiderMenu collapsed={siderCollapsed} collapsible onCollapse={(collapsed) => setSiderCollapsed(collapsed)} />
            </div>
            <div className="w-full px-2 pb-2">
              <Menu
                style={{ background: "transparent", borderInlineEnd: "none" }}
                inlineCollapsed={siderCollapsed}
                items={[
                  {
                    key: "$expand",
                    icon: (
                      <span className="anticon scale-125" role="img">
                        <IconLayoutSidebarRightCollapse size="1em" />
                      </span>
                    ),
                    label: t("common.menu.expand_sidebar"),
                    onClick: () => setSiderCollapsed(false),
                  },
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
                ].filter((item) => {
                  if (item.key === "$expand") {
                    return siderCollapsed;
                  }
                  return true;
                })}
                mode="inline"
                selectable={false}
              />
            </div>
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
                className="size-full"
                style={{
                  backgroundImage:
                    "linear-gradient(rgba(255, 255, 255, 0.063) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.063) 1px, transparent 1px)",
                  backgroundSize: "20px 20px",
                }}
              >
                <div className="size-full backdrop-blur-[1px]"></div>
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

const SiderMenu = ({
  collapsed,
  collapsible,
  onSelect,
  onCollapse,
}: {
  collapsed?: boolean;
  collapsible?: boolean;
  onSelect?: (key: string) => void;
  onCollapse?: (collapsed: boolean) => void;
}) => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const MENU_KEY_HOME = "/";
  const MENU_KEY_WORKFLOWS = "/workflows";
  const MENU_KEY_CERTIFICATES = "/certificates";
  const MENU_KEY_ACME = "/acme";
  const MENU_KEY_ACCESSES = "/accesses";
  const MENU_KEY_PRESETS = "/presets";
  const MENU_KEY_PLUGINS = "/plugins";
  const MENU_KEY_SETTINGS = "/settings";
  const menuItems: Required<MenuProps>["items"] = (() => {
    const convert = (key: string, label: string, icon?: React.ReactNode) => {
      return {
        key: key,
        icon: icon ? (
          <span className="anticon scale-125" role="img">
            {icon}
          </span>
        ) : (
          void 0
        ),
        label: t(label),
        onClick: () => {
          if (!key || key.startsWith("$")) return;

          navigate(key);
          onSelect?.(key);
        },
      };
    };

    const items = [
      convert(MENU_KEY_HOME, "dashboard.page.heading", <IconHome size="1em" />),
      convert(MENU_KEY_WORKFLOWS, "workflow.page.heading", <IconHierarchy3 size="1em" />),
      convert(MENU_KEY_CERTIFICATES, "certificate.page.heading", <IconCertificate size="1em" />),
      {
        ...convert(MENU_KEY_ACME, "acme.page.heading", <IconRobot size="1em" />),
        children: [
          convert(MENU_KEY_ACME + "/ca", "acme.ca.heading"),
          convert(MENU_KEY_ACME + "/accounts", "acme.accounts.heading"),
          convert(MENU_KEY_ACME + "/configs", "acme.configs.heading"),
        ],
        onClick: void 0,
      },
      convert(MENU_KEY_ACCESSES, "access.page.heading", <IconPlugConnected size="1em" />),
      convert(MENU_KEY_PRESETS, "preset.page.heading", <IconCodeDots size="1em" />),
      convert(MENU_KEY_PLUGINS, "plugin.page.heading", <IconPuzzle size="1em" />),
      convert(MENU_KEY_SETTINGS, "settings.page.heading", <IconSettings size="1em" />),
    ];

    return items;
  })();
  const menuItemsFlatten = menuItems.flatMap((item) => [item, ...("children" in item! ? (item.children ?? []) : [])]);
  const menuSelectedKeys = (() => {
    const selected =
      menuItemsFlatten.find((item) => item!.key === location.pathname) ??
      menuItemsFlatten.find((item) => item!.key !== MENU_KEY_HOME && location.pathname.startsWith(item!.key as string));
    return selected ? [selected.key as string] : [];
  })();
  const menuDefaultOpenKeys = [MENU_KEY_ACME];

  useEffect(() => {
    void usePluginCatalogStore.getState().reload();
  }, []);

  return (
    <>
      <div className="h-[72px] w-full overflow-hidden px-4 py-2 max-md:py-0">
        <div className="flex size-full items-center justify-around space-x-4">
          <img src={withBasePath("/logo.svg")} className={mergeCls("transition-all duration-300", collapsed ? "size-[36px]" : "size-[42px]")} />
          <div className="flex-1">
            <Show when={!collapsed}>
              <div className="truncate text-base font-semibold">Certimate</div>
              <AppVersion.LinkButton className="text-xs" />
            </Show>
          </div>
          {collapsible && (
            <div className="relative -right-2">
              <Tooltip title={t("common.menu.collapse_sidebar")} mouseEnterDelay={1} placement="right">
                <Button
                  className="bg-background shadow-sm"
                  icon={<IconLayoutSidebarLeftCollapse size="1.25em" stroke="1.25" color="#999" />}
                  shape="circle"
                  size="small"
                  type="text"
                  onClick={() => onCollapse?.(true)}
                />
              </Tooltip>
            </div>
          )}
        </div>
      </div>
      <div className="w-full grow overflow-x-hidden overflow-y-auto">
        <Menu
          style={{ background: "transparent", borderInlineEnd: "none" }}
          defaultOpenKeys={menuDefaultOpenKeys}
          forceSubMenuRender
          inlineCollapsed={collapsed}
          items={menuItems}
          mode="inline"
          selectedKeys={menuSelectedKeys}
        />
      </div>
    </>
  );
};

const SiderMenuDrawer = ({ trigger }: { trigger: React.ReactNode }) => {
  const { token: themeToken } = theme.useToken();

  const [siderOpen, setSiderOpen] = useState(false);

  const triggerEl = useTriggerElement(trigger, { onClick: () => setSiderOpen(true) });

  const handleMenuSelect = () => {
    setSiderOpen(false);
  };

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
};

export default ConsoleLayout;
