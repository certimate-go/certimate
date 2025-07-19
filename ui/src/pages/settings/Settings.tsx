import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { IconDatabaseCog, IconPalette, IconPlugConnected, IconUserShield } from "@tabler/icons-react";
import { Menu } from "antd";

const Settings = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const menus = [
    ["account", "settings.account.tab", <IconUserShield size="1em" />],
    ["appearance", "settings.appearance.tab", <IconPalette size="1em" />],
    ["ssl-provider", "settings.sslprovider.tab", <IconPlugConnected size="1em" />],
    ["persistence", "settings.persistence.tab", <IconDatabaseCog size="1em" />],
  ] satisfies [string, string, React.ReactElement][];
  const [menuKey, setMenuValue] = useState("account");
  useEffect(() => {
    const path = location.pathname.split("/")[2];
    if (!path) {
      navigate("/settings/account");
      return;
    }

    setMenuValue(path);
  }, [location.pathname]);

  return (
    <div className="px-6 py-4">
      <div className="mx-auto max-w-320">
        <h1>{t("settings.page.title")}</h1>

        <div className="hidden select-none max-lg:block">
          <Menu
            style={{ background: "transparent", borderInlineEnd: "none" }}
            mode="horizontal"
            selectedKeys={[menuKey]}
            items={menus.map(([key, label, icon]) => ({
              key,
              label: t(label),
              icon: <span className="anticon scale-125">{icon}</span>,
            }))}
            onClick={({ key }) => {
              setMenuValue(key);
              navigate(`/settings/${key}`);
            }}
          />
        </div>

        <div className="flex h-full justify-stretch gap-x-4 overflow-hidden">
          <div className="my-[6px] w-[256px] select-none max-lg:hidden">
            <Menu
              style={{ background: "transparent", borderInlineEnd: "none" }}
              mode="vertical"
              selectedKeys={[menuKey]}
              items={menus.map(([key, label, icon]) => ({
                key,
                label: t(label),
                icon: <span className="anticon scale-125">{icon}</span>,
              }))}
              onClick={({ key }) => {
                setMenuValue(key);
                navigate(`/settings/${key}`);
              }}
            />
          </div>

          <div className="flex-1">
            <div className="px-4 max-lg:px-0 max-lg:py-6">
              <Outlet />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;
