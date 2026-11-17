import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, useLocation } from "react-router-dom";
import { IconRefresh, IconServerOff } from "@tabler/icons-react";
import { useRequest } from "ahooks";
import { Button, Card, Spin, Typography, theme } from "antd";

import AuthBackdrop from "@/components/AuthBackdrop";
import AuthFooter from "@/components/AuthFooter";
import { getAuthStore } from "@/repository/admin";
import { getSetupStatus } from "@/repository/setup";
import { unwrapErrMsg } from "@/utils/error";
import { withBasePath } from "@/utils/url";

type SetupGateMode = "app" | "setup";

type SetupGateProps = {
  mode?: SetupGateMode;
  children?: React.ReactNode;
};

const SetupGate = ({ mode = "app", children }: SetupGateProps) => {
  const { t } = useTranslation();

  const location = useLocation();

  const { token: themeToken } = theme.useToken();

  const { data, loading, error, run } = useRequest(getSetupStatus, {
    refreshDeps: [location.pathname],
  });

  const initialized = data?.initialized;

  useEffect(() => {
    if (initialized === false) {
      const auth = getAuthStore();
      if (auth.isValid) {
        auth.clear();
      }
    }
  }, [initialized]);

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

  if (error || initialized == null) {
    return (
      <>
        <AuthBackdrop />
        <div className="flex h-screen w-full items-center justify-center p-4">
          <Card className="w-120 max-w-full rounded-md shadow-md max-sm:size-full max-sm:rounded-none">
            <div className="px-4 py-8">
              <div className="mb-8 text-center" role="alert">
                <div className="mb-6 flex justify-center">
                  <div
                    className="flex size-16 items-center justify-center rounded-full"
                    style={{ backgroundColor: themeToken.colorErrorBg }}
                  >
                    <IconServerOff size={28} stroke={1.5} style={{ color: themeToken.colorError }} />
                  </div>
                </div>
                <Typography.Title className="mb-2" level={4}>
                  {t("login.setup.gate.title")}
                </Typography.Title>
                <Typography.Text className="text-sm" type="secondary">
                  {t("login.setup.gate.error")}
                </Typography.Text>
                {error != null && (
                  <div className="mt-2 text-xs">
                    <Typography.Text type="secondary">{unwrapErrMsg(error)}</Typography.Text>
                  </div>
                )}
              </div>
              <Button
                block
                icon={<IconRefresh size="1.25em" />}
                loading={loading}
                size="large"
                type="primary"
                onClick={run}
              >
                {t("login.setup.gate.retry")}
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

  if (!initialized) {
    if (mode === "app") {
      return <Navigate replace to="/setup" />;
    }
    return children;
  }

  if (mode === "setup") {
    return <Navigate replace to="/login" />;
  }
  return children;
};

export default SetupGate;
