import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { IconArrowRight, IconLock, IconMail } from "@tabler/icons-react";
import { App, Button, Card, Form, Input, Space } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AuthBackdrop from "@/components/AuthBackdrop";
import AuthFooter from "@/components/AuthFooter";
import { useAntdForm } from "@/hooks";
import { authWithPassword } from "@/repository/admin";
import { unwrapErrMsg } from "@/utils/error";
import { withBasePath } from "@/utils/url";

const Login = () => {
  const navigage = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const formSchema = z.object({
    username: z.email(),
    password: z.string().min(10).max(72),
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
            <div className="mb-12 flex items-center justify-center">
              <img src={withBasePath("/logo.svg")} className="w-16" />
            </div>

            <Form {...formProps} form={formInst} disabled={formPending} layout="vertical" validateTrigger="onBlur">
              <Form.Item name="username" label={t("login.form.username.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconMail size="1.25em" />
                  </Space.Addon>
                  <Input autoComplete="new-password" autoFocus placeholder={t("login.form.username.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item name="password" label={t("login.form.password.label")} rules={[formRule]}>
                <Space.Compact block>
                  <Space.Addon>
                    <IconLock size="1.25em" />
                  </Space.Addon>
                  <Input.Password autoComplete="new-password" placeholder={t("login.form.password.placeholder")} size="large" />
                </Space.Compact>
              </Form.Item>

              <Form.Item className="mt-8 mb-0">
                <Button block type="primary" htmlType="submit" icon={<IconArrowRight size="1.25em" />} iconPlacement="end" loading={formPending} size="large">
                  {t("login.form.submit.button")}
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

export default Login;
