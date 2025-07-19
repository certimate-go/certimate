import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { App, Button, Divider, Flex, Form, Input, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { useAntdForm } from "@/hooks";
import { authWithPassword, getAuthStore, save as saveAdmin } from "@/repository/admin";
import { getErrMsg } from "@/utils/error";

const SettingsAccount = () => {
  const { t } = useTranslation();

  return (
    <>
      <h2>{t("settings.account.username.title")}</h2>
      <div className="md:max-w-160">
        <SettingsAccountUsername />
      </div>

      <Divider />

      <h2>{t("settings.account.password.title")}</h2>
      <div className="md:max-w-160">
        <SettingsAccountPassword />
      </div>

      <Divider />

      <h2>{t("settings.account.2fa.title")}</h2>
      <div>TODO ...</div>
    </>
  );
};

const SettingsAccountUsername = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const formSchema = z.object({
    username: z.email(t("common.errmsg.email_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      username: getAuthStore().record?.email,
    },
    onSubmit: async (values) => {
      try {
        await saveAdmin({ email: values.username });

        message.success(t("common.text.operation_succeeded"));

        setTimeout(() => {
          getAuthStore().clear();
          navigate("/login");
        }, 500);
      } catch (err) {
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });
  const [formVisible, setFormVisible] = useState(false);
  const [formChanged, setFormChanged] = useState(false);

  const handleInputChange = () => {
    setFormChanged(formInst.getFieldValue("username") !== formProps.initialValues?.username);
  };

  const handleEditClick = () => {
    setFormVisible(true);
    formInst.resetFields();
  };

  const handleCancelClick = () => {
    setFormVisible(false);
    setFormChanged(false);
  };

  return (
    <div className={className} style={style}>
      <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
        {formVisible ? (
          <>
            <Form.Item name="username" label={t("settings.account.username.form.email.label")} rules={[formRule]}>
              <Input autoFocus placeholder={t("settings.account.username.form.email.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item>
              <Flex align="center" gap="small">
                <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
                  {t("common.button.save")}
                </Button>
                <Button disabled={formPending} onClick={handleCancelClick}>
                  {t("common.button.cancel")}
                </Button>
              </Flex>
            </Form.Item>
          </>
        ) : (
          <>
            <div className="mb-2">
              <Typography.Text type="secondary">{t("settings.account.username.tips")}</Typography.Text>
            </div>
            <Flex align="center" gap="small">
              <Typography.Text>{getAuthStore().record?.email}</Typography.Text>
              <Button onClick={handleEditClick}>{t("settings.account.username.button.label")}</Button>
            </Flex>
          </>
        )}
      </Form>
    </div>
  );
};

const SettingsAccountPassword = ({ className, style }: { className?: string; style?: React.CSSProperties }) => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const formSchema = z.object({
    oldPassword: z.string().min(10, t("settings.account.password.form.email.password.errmsg.invalid")),
    newPassword: z.string().min(10, t("settings.account.password.form.email.password.errmsg.invalid")),
    confirmPassword: z.string().refine((v) => v === formInst.getFieldValue("newPassword"), {
      error: t("settings.account.password.form.email.password.errmsg.not_matched"),
    }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
  } = useAntdForm({
    initialValues: {
      oldPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
    onSubmit: async (values) => {
      try {
        await authWithPassword(getAuthStore().record!.email, values.oldPassword);
        await saveAdmin({ password: values.newPassword, passwordConfirm: values.confirmPassword });

        message.success(t("common.text.operation_succeeded"));

        setTimeout(() => {
          getAuthStore().clear();
          navigate("/login");
        }, 500);
      } catch (err) {
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });
  const [formVisible, setFormVisible] = useState(false);
  const [formChanged, setFormChanged] = useState(false);

  const handleInputChange = () => {
    const values = formInst.getFieldsValue();
    setFormChanged(!!values.oldPassword && !!values.newPassword && !!values.confirmPassword);
  };

  const handleEditClick = () => {
    setFormVisible(true);
    formInst.resetFields();
  };

  const handleCancelClick = () => {
    setFormVisible(false);
    setFormChanged(false);
  };

  return (
    <div className={className} style={style}>
      <Form {...formProps} form={formInst} disabled={formPending} layout="vertical">
        {formVisible ? (
          <>
            <Form.Item name="oldPassword" label={t("settings.account.password.form.email.old_password.label")} rules={[formRule]}>
              <Input.Password autoFocus placeholder={t("settings.account.password.form.email.old_password.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item name="newPassword" label={t("settings.account.password.form.email.new_password.label")} rules={[formRule]}>
              <Input.Password placeholder={t("settings.account.password.form.email.new_password.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item name="confirmPassword" label={t("settings.account.password.form.email.confirm_password.label")} rules={[formRule]}>
              <Input.Password placeholder={t("settings.account.password.form.email.confirm_password.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item>
              <Flex align="center" gap="small">
                <Button type="primary" htmlType="submit" disabled={!formChanged} loading={formPending}>
                  {t("common.button.save")}
                </Button>
                <Button disabled={formPending} onClick={handleCancelClick}>
                  {t("common.button.cancel")}
                </Button>
              </Flex>
            </Form.Item>
          </>
        ) : (
          <>
            <div className="mb-2">
              <Typography.Text type="secondary">{t("settings.account.password.tips")}</Typography.Text>
            </div>

            <Button onClick={handleEditClick}>{t("settings.account.password.button.label")}</Button>
          </>
        )}
      </Form>
    </div>
  );
};

export default SettingsAccount;
