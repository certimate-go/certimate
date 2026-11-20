import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { IconArrowRight, IconKey, IconLock, IconMail } from "@tabler/icons-react";
import { App, Button, Card, Form, Input, Space, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { getStatus as getInstallerStatus, initSuperuser } from "@/api/installer";
import AuthBackdrop from "@/components/AuthBackdrop";
import AuthFooter from "@/components/AuthFooter";
import { useAntdForm } from "@/hooks";
import { unwrapErrMsg } from "@/utils/error";
import { withBasePath } from "@/utils/url";

const Setup = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const formSchema = z
    .object({
      token: z.string().length(32),
      username: z.email(),
      password: z.string().min(10).max(72),
      confirmPassword: z.string().min(10).max(72),
    })
    .superRefine((values, ctx) => {
      if (values.confirmPassword !== values.password) {
        ctx.addIssue({
          code: "custom",
          message: t("installer.form.confirm_password.invalid.not_matched"),
          path: ["confirmPassword"],
        });
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
      try {
        await initSuperuser({
          token: values.token,
          email: values.username,
          password: values.password,
        });
        await getInstallerStatus();
        await navigate("/login");
      } catch (err) {
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

        throw err;
      }
    },
  });

  return (
    <>
      <AuthBackdrop />

      <div className="flex h-screen w-full flex-col items-center justify-center">
        <Card className="w-120 max-w-full rounded-md shadow-md max-sm:size-full max-sm:rounded-none">
          <div className="px-4 py-8">
            <div className="mb-6 flex items-center justify-center">
              <img src={withBasePath("/logo.svg")} className="w-16" />
            </div>

            <div className="mb-10 text-center">
              <Typography.Title className="mb-2" level={4}>
                {t("installer.page.heading")}
              </Typography.Title>
              <Typography.Text className="text-sm" type="secondary">
                {t("installer.page.subheading")}
              </Typography.Text>
            </div>

            <Form {...formProps} form={formInst} disabled={formPending} layout="vertical" validateTrigger="onBlur">
              <Form.Item name="token" label={t("installer.form.token.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconKey size="1.25em" />
                  </Space.Addon>
                  <Input autoComplete="off" autoFocus placeholder={t("installer.form.token.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="username" label={t("installer.form.username.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconMail size="1.25em" />
                  </Space.Addon>
                  <Input autoComplete="off" placeholder={t("installer.form.username.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="password" label={t("installer.form.password.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconLock size="1.25em" />
                  </Space.Addon>
                  <Input.Password autoComplete="new-password" placeholder={t("installer.form.password.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="confirmPassword" label={t("installer.form.confirm_password.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconLock size="1.25em" />
                  </Space.Addon>
                  <Input.Password autoComplete="new-password" placeholder={t("installer.form.confirm_password.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item className="mt-8 mb-0">
                <Button block type="primary" htmlType="submit" icon={<IconArrowRight size="1.25em" />} iconPlacement="end" loading={formPending} size="large">
                  {t("installer.form.submit.button")}
                </Button>
              </Form.Item>
            </Form>

            <div className="mt-12">
              <AuthFooter />
            </div>
          </div>
        </Card>
      </div>
    </>
  );
};

export default Setup;
