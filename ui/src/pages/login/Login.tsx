import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { IconArrowRight, IconLock, IconMail } from "@tabler/icons-react";
import { App, Button, Card, Divider, Form, Input, Space } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import AppDocument from "@/components/AppDocument";
import AppLocale from "@/components/AppLocale";
import AppTheme from "@/components/AppTheme";
import AppVersion from "@/components/AppVersion";
import { useAntdForm, useBrowserTheme } from "@/hooks";

import { authWithPassword } from "@/repository/admin";
import { getErrMsg } from "@/utils/error";

const Login = () => {
  const navigage = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();
  const { theme: browserTheme } = useBrowserTheme();

  const bgStyle = useMemo<React.CSSProperties>(() => {
    let svg = "";
    let mask = "";
    if (browserTheme === "dark") {
      svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="32" height="32" fill="none" stroke="rgb(202 78 13 / 0.12)"><path d="M0 .5H31.5V32"/></svg>`;
      mask = "white";
    } else {
      svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="32" height="32" fill="none" stroke="rgb(249 115 22 / 0.08)"><path d="M0 .5H31.5V32"/></svg>`;
      mask = "black";
    }

    return {
      backgroundImage: `url('data:image/svg+xml;base64,${btoa(svg)}')`,
      maskImage: `linear-gradient(to bottom right, transparent, ${mask}, transparent)`,
    };
  }, [browserTheme]);

  const formSchema = z.object({
    username: z.email(t("login.username.errmsg.invalid")),
    password: z.string().min(10, t("login.password.errmsg.invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      username: "",
      password: "",
    },
    onSubmit: async (values) => {
      try {
        await authWithPassword(values.username, values.password);
        await navigage("/");
      } catch (err) {
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });

  return (
    <>
      <div className="pointer-events-none fixed min-h-screen w-full" style={bgStyle}></div>

      <div className="flex h-screen w-full flex-col items-center justify-center">
        <Card className="w-120 max-w-full rounded-md shadow-md max-sm:h-full max-sm:w-full max-sm:rounded-none">
          <div className="px-4 py-8">
            <div className="mb-12 flex items-center justify-center">
              <img src="/logo.svg" className="w-16" />
            </div>

            <Form {...formProps} form={formInst} disabled={formPending} layout="vertical" validateTrigger="onBlur">
              <Form.Item name="username" label={t("login.username.label")} rules={[formRule]}>
                <Input
                  addonBefore={<IconMail size="1.25em" />}
                  autoComplete="new-password"
                  autoFocus
                  placeholder={t("login.username.placeholder")}
                  size="large"
                />
              </Form.Item>

              <Form.Item name="password" label={t("login.password.label")} rules={[formRule]}>
                <Input.Password
                  addonBefore={<IconLock size="1.25em" />}
                  autoComplete="new-password"
                  placeholder={t("login.password.placeholder")}
                  size="large"
                />
              </Form.Item>

              <Form.Item className="mt-8 mb-0">
                <Button block type="primary" htmlType="submit" icon={<IconArrowRight size="1.25em" />} iconPosition="end" loading={formPending} size="large">
                  {t("login.submit")}
                </Button>
              </Form.Item>
            </Form>

            <div className="mt-12">
              <div className="block max-sm:hidden">
                <div className="flex items-center justify-center">
                  <Space align="center" split={<Divider type="vertical" />} size={4}>
                    <AppLocale.LinkButton />
                    <AppTheme.LinkButton />
                    <AppDocument.LinkButton />
                    <AppVersion.LinkButton />
                  </Space>
                </div>
              </div>
              <div className="hidden max-sm:block">
                <div className="flex items-center justify-center">
                  <Space align="center" split={<Divider type="vertical" />} size={4}>
                    <AppLocale.LinkButton />
                    <AppTheme.LinkButton />
                    <AppDocument.LinkButton />
                  </Space>
                </div>
                <div className="mt-6 flex items-center justify-center">
                  <Space align="center" split={<Divider type="vertical" />} size={4}>
                    <AppVersion.LinkButton />
                  </Space>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </div>
    </>
  );
};

export default Login;
