import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Button, Card, Form, Input, notification } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { useAntdForm } from "@/hooks";
import { authWithPassword } from "@/repository/admin";
import { getErrMsg } from "@/utils/error";

const Login = () => {
  const navigage = useNavigate();

  const { t } = useTranslation();

  const [notificationApi, NotificationContextHolder] = notification.useNotification();

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
    onSubmit: async (values) => {
      try {
        await authWithPassword(values.username, values.password);
        await navigage("/");
      } catch (err) {
        notificationApi.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });

  return (
    <>
      {NotificationContextHolder}

      <Card className="mx-auto mt-32 max-w-[35em] rounded-md border p-10 shadow-md dark:border-stone-500">
        <div className="mb-10 flex items-center justify-center">
          <img src="/logo.svg" className="w-16" />
        </div>

        <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
          <Form.Item name="username" label={t("login.username.label")} rules={[formRule]}>
            <Input placeholder={t("login.username.placeholder")} />
          </Form.Item>

          <Form.Item name="password" label={t("login.password.label")} rules={[formRule]}>
            <Input type="password" placeholder={t("login.password.placeholder")} />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={formPending}>
              {t("login.submit")}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
};

export default Login;
