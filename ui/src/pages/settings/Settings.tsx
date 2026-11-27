import { useTranslation } from "react-i18next";
import { Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import { IconDatabaseCog, IconHeartRateMonitor, IconInfoCircle, IconPalette, IconRobot, IconShieldLock } from "@tabler/icons-react";
import { Card, Menu, Tabs } from "antd";

const Settings = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const subpath = location.pathname.split("/")[2];

  const menus = [
    ["account", "settings.security.tab", <IconShieldLock size="1em" />],
    ["appearance", "settings.appearance.tab", <IconPalette size="1em" />],
    ["ssl-provider", "settings.sslprovider.tab", <IconRobot size="1em" />],
    ["persistence", "settings.persistence.tab", <IconDatabaseCog size="1em" />],
    ["diagnostics", "settings.diagnostics.tab", <IconHeartRateMonitor size="1em" />],
    ["about", "settings.about.tab", <IconInfoCircle size="1em" />],
  ] satisfies [string, string, React.ReactElement<unknown>][];
  const menuKey = subpath;

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(`/settings/${key}`);
  };

  if (!subpath) {
    return <Navigate to="/settings/account" replace />;
  }

  return (
    <div className="px-6 py-4">
      <div className="container mb-6">
        <h1 className="text-2xl font-medium text-foreground/80">{t("settings.page.heading")}</h1>
        <p className="mt-1 text-sm text-foreground/70">{t("settings.page.subheading")}</p>
      </div>

      <div className="container">
        <div className="block lg:hidden">
          <Tabs
            className="-mt-2"
            styles={{
              header: {
                paddingLeft: "2px",
                paddingRight: "2px",
                marginBottom: "0",
              },
            }}
            activeKey={menuKey}
            items={menus.map(([key, label, icon]) => ({
              key,
              label: t(label),
              icon: (
                <span className="anticon scale-125" role="img">
                  {icon}
                </span>
              ),
            }))}
            size="large"
            onChange={(key) => handleMenuClick({ key })}
          />
        </div>

        <div className="mt-4 grid grid-cols-1 gap-6 lg:mt-0 lg:grid-cols-[1fr_280px]">
          <div className="min-w-0 space-y-6 overflow-hidden">
            <Outlet />
          </div>

          <aside className="hidden lg:block">
            <Card size="small">
              <Menu
                style={{ background: "transparent", borderInlineEnd: "none" }}
                mode="vertical"
                selectedKeys={[menuKey]}
                items={menus.map(([key, label, icon]) => ({
                  key,
                  label: t(label),
                  icon: (
                    <span className="anticon scale-125" role="img">
                      {icon}
                    </span>
                  ),
                }))}
                onClick={handleMenuClick}
              />
            </Card>
          </aside>
        </div>
      </div>
    </div>
  );
};

export default Settings;
