import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { IconArrowRight, IconKey, IconLock, IconMail } from "@tabler/icons-react";
import { App, Button, Card, Divider, Form, Input, Space, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AppDocument from "@/components/AppDocument";
import AppLocale from "@/components/AppLocale";
import AppTheme from "@/components/AppTheme";
import AppVersion from "@/components/AppVersion";
import { useAntdForm, useBrowserTheme } from "@/hooks";
import { getSetupStatus, submitSetup } from "@/repository/setup";
import { unwrapErrMsg } from "@/utils/error";
import { withBasePath } from "@/utils/url";

const Setup = () => {
  const navigate = useNavigate();

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

  const isValidPassword = (value: string) => {
    const codePoints = Array.from(value).length;
    const bytes = new TextEncoder().encode(value).length;
    return codePoints >= 10 && codePoints <= 71 && bytes <= 72;
  };

  const formSchema = z
    .object({
      token: z.string().min(1),
      username: z.email(t("login.setup.form.username.invalid")),
      password: z.string().refine(isValidPassword, t("login.setup.form.password.invalid")),
      confirmPassword: z.string(),
    })
    .superRefine((values, ctx) => {
      if (values.confirmPassword !== values.password) {
        ctx.addIssue({ code: "custom", path: ["confirmPassword"], message: t("login.setup.form.confirm_password.invalid") });
      }
    });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      token: "",
      username: "",
      password: "",
      confirmPassword: "",
    },
    onSubmit: async (values) => {
      const cleanup = () => {
        formInst.setFieldsValue({ token: "", password: "", confirmPassword: "" });
      };
      try {
        await submitSetup({
          token: values.token,
          email: values.username,
          password: values.password,
          passwordConfirm: values.confirmPassword,
        });
      } catch (err) {
        try {
          const status = await getSetupStatus();
          if (status.initialized) {
            cleanup();
            await navigate("/login");
            return;
          }
        } catch {
          // fall through to the error notification
        }

        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

        throw err;
      }
      cleanup();
      await navigate("/login");
    },
  });

  return (
    <>
      <div className="pointer-events-none fixed min-h-screen w-full" style={bgStyle}></div>

      <div className="flex h-screen w-full flex-col items-center justify-center">
        <Card className="w-120 max-w-full rounded-md shadow-md max-sm:size-full max-sm:rounded-none">
          <div className="px-4 py-8">
            <div className="mb-6 flex items-center justify-center">
              <img src={withBasePath("/logo.svg")} className="w-16" />
            </div>

            <div className="mb-10 text-center">
              <Typography.Title className="mb-2" level={4}>
                {t("login.setup.heading")}
              </Typography.Title>
              <Typography.Text className="text-sm" type="secondary">
                {t("login.setup.description")}
              </Typography.Text>
            </div>

            <Form {...formProps} form={formInst} disabled={formPending} layout="vertical" validateTrigger="onBlur">
              <Form.Item name="token" label={t("login.setup.form.token.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconKey size="1.25em" />
                  </Space.Addon>
                  <Input autoComplete="off" autoFocus placeholder={t("login.setup.form.token.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="username" label={t("login.setup.form.username.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconMail size="1.25em" />
                  </Space.Addon>
                  <Input autoComplete="off" placeholder={t("login.setup.form.username.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="password" label={t("login.setup.form.password.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconLock size="1.25em" />
                  </Space.Addon>
                  <Input.Password autoComplete="new-password" placeholder={t("login.setup.form.password.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="confirmPassword" label={t("login.setup.form.confirm_password.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconLock size="1.25em" />
                  </Space.Addon>
                  <Input.Password autoComplete="new-password" placeholder={t("login.setup.form.confirm_password.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item className="mt-8 mb-0">
                <Button block type="primary" htmlType="submit" icon={<IconArrowRight size="1.25em" />} iconPlacement="end" loading={formPending} size="large">
                  {t("login.setup.form.submit.button")}
                </Button>
              </Form.Item>
            </Form>

            <div className="mt-12">
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
            </div>
          </div>
        </Card>
      </div>
    </>
  );
};

export default Setup;
