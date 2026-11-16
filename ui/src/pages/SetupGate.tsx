import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, useLocation } from "react-router-dom";
import { useRequest } from "ahooks";
import { Alert, Button, Card, Space, Spin } from "antd";

import { getAuthStore } from "@/repository/admin";
import { getSetupStatus } from "@/repository/setup";

type SetupGateMode = "app" | "setup";

type SetupGateProps = {
  mode?: SetupGateMode;
  children?: React.ReactNode;
};

const SetupGate = ({ mode = "app", children }: SetupGateProps) => {
  const { t } = useTranslation();

  const location = useLocation();

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

  if (loading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Spin size="large" />
      </div>
    );
  }

  if (error || initialized == null) {
    return (
      <div className="flex h-screen w-full items-center justify-center p-4">
        <Card className="max-w-lg">
          <Space align="start" direction="vertical" size="middle">
            <Alert showIcon title={t("login.setup.gate.error")} type="error" />
            <Button type="primary" onClick={run}>
              {t("login.setup.gate.retry")}
            </Button>
          </Space>
        </Card>
      </div>
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
