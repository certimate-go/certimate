import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, useLocation } from "react-router-dom";
import { IconRefresh, IconServerOff } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { Button, Card, Spin, Typography, theme } from "antd";

import { getStatus as getInstallerStatus } from "@/api/installer";
import AuthBackdrop from "@/components/AuthBackdrop";
import AuthFooter from "@/components/AuthFooter";
import { getAuthStore } from "@/repository/admin";
import { unwrapErrMsg } from "@/utils/error";
import { withBasePath } from "@/utils/url";

type InstallerGuardProps = {
  children?: React.ReactNode;
};

const InstallerGuard = ({ children }: InstallerGuardProps) => {
  const location = useLocation();

  const { t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { data, loading, error, refresh } = useRequest(
    async () => {
      const res = await getInstallerStatus();
      return res.data;
    },
    {
      refreshDeps: [location.pathname],
    }
  );

  const installed = data?.installed;
  useEffect(() => {
    if (installed === false) {
      const auth = getAuthStore();
      if (auth.isValid) {
        auth.clear();
      }
    }
  }, [installed]);

  const handleClickReload = () => {
    refresh();
  };

  if (loading && data == null && error == null) {
    return (
      <>
        <AuthBackdrop />
        <div className="flex h-screen w-full flex-col items-center justify-center">
          <img src={withBasePath("/logo.svg")} className="w-16" />
          <Spin className="mt-8" size="large" />
        </div>
      </>
    );
  }

  if (error || installed == null) {
    return (
      <>
        <AuthBackdrop />
        <div className="flex h-screen w-full items-center justify-center p-4">
          <Card className="w-120 max-w-full rounded-md shadow-md max-sm:size-full max-sm:rounded-none">
            <div className="px-4 py-8">
              <div className="mb-8 text-center" role="alert">
                <div className="mb-6 flex justify-center">
                  <div className="flex size-16 items-center justify-center rounded-full" style={{ backgroundColor: themeToken.colorErrorBg }}>
                    <IconServerOff size={28} stroke={1.5} style={{ color: themeToken.colorError }} />
                  </div>
                </div>
                <Typography.Title className="mb-2" level={4}>
                  {t("installer.guard.title")}
                </Typography.Title>
                {error != null && (
                  <div className="mt-2 text-xs">
                    <Typography.Text type="secondary">{unwrapErrMsg(error)}</Typography.Text>
                  </div>
                )}
              </div>
              <Button block icon={<IconRefresh size="1.25em" />} loading={loading} size="large" type="primary" onClick={handleClickReload}>
                {t("common.button.reload")}
              </Button>
              <div className="mt-12">
                <AuthFooter />
              </div>
            </div>
          </Card>
        </div>
      </>
    );
  }

  if (installed) {
    if (location.pathname !== "/login") {
      return <Navigate replace to="/login" />;
    }
    return children;
  } else {
    if (location.pathname !== "/installer") {
      return <Navigate replace to="/installer" />;
    }
    return children;
  }
};

export default InstallerGuard;
