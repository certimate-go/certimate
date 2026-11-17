import { Divider, Space } from "antd";

import AppDocument from "@/components/AppDocument";
import AppLocale from "@/components/AppLocale";
import AppTheme from "@/components/AppTheme";
import AppVersion from "@/components/AppVersion";

const AuthFooter = () => {
  return (
    <>
      <div className="block max-sm:hidden">
        <div className="flex items-center justify-center">
          <Space align="center" separator={<Divider orientation="vertical" />} size={4}>
            <AppLocale.LinkButton />
            <AppTheme.LinkButton />
            <AppDocument.LinkButton />
            <AppVersion.LinkButton />
          </Space>
        </div>
      </div>
      <div className="hidden max-sm:block">
        <div className="flex items-center justify-center">
          <Space align="center" separator={<Divider orientation="vertical" />} size={4}>
            <AppLocale.LinkButton />
            <AppTheme.LinkButton />
            <AppDocument.LinkButton />
          </Space>
        </div>
        <div className="mt-6 flex items-center justify-center">
          <Space align="center" separator={<Divider orientation="vertical" />} size={4}>
            <AppVersion.LinkButton />
          </Space>
        </div>
      </div>
    </>
  );
};

export default AuthFooter;
