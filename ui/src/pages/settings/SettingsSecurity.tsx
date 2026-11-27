import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { App, Button, Card, Flex, Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useAntdForm } from "@/hooks";
import { authWithPassword, getAuthStore, save as saveAdmin } from "@/repository/admin";
import { unwrapErrMsg } from "@/utils/error";

const SettingsSecurity = () => {
  return (
    <>
      <div>
        <SettingsSecurityUsername />
      </div>

      <div>
        <SettingsSecurityPassword />
      </div>
    </>
  );
};

const SettingsSecurityUsername = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const formSchema = z.object({
    username: z.email(),
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
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

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
    <Card
      title={
        <div className="mt-4 mb-0 overflow-hidden">
          <h2 className="truncate text-base font-semibold">{t("settings.security.username.title")}</h2>
          <p className="text-sm font-normal whitespace-normal text-foreground/70">{t("settings.security.username.subtitle")}</p>
        </div>
      }
    >
      <Form className="md:max-w-160" {...formProps} form={formInst} disabled={formPending} layout="vertical">
        {formVisible ? (
          <>
            <Form.Item name="username" label={t("settings.security.username.form.email.label")} rules={[formRule]}>
              <Input autoFocus placeholder={t("settings.security.username.form.email.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item noStyle>
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
            <Form.Item label={t("settings.security.username.form.email.label")}>
              <Input value={getAuthStore().record?.email} placeholder="" variant="filled" onChange={handleInputChange} />
            </Form.Item>

            <Form.Item noStyle>
              <Button onClick={handleEditClick}>{t("settings.security.username.button")}</Button>
            </Form.Item>
          </>
        )}
      </Form>
    </Card>
  );
};

const SettingsSecurityPassword = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const formSchema = z
    .object({
      oldPassword: z.string().min(10).max(72),
      newPassword: z.string().min(10).max(72),
      confirmPassword: z.string().min(10).max(72),
    })
    .refine((values) => values.newPassword === values.confirmPassword, {
      error: t("settings.security.password.form.password.errmsg.not_matched"),
      path: ["confirmPassword"],
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
        notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

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
    <Card
      title={
        <div className="mt-4 mb-0 overflow-hidden">
          <h2 className="truncate text-base font-semibold">{t("settings.security.password.title")}</h2>
          <p className="text-sm font-normal whitespace-normal text-foreground/70">{t("settings.security.password.subtitle")}</p>
        </div>
      }
    >
      <Form className="md:max-w-160" {...formProps} form={formInst} disabled={formPending} layout="vertical">
        {formVisible ? (
          <>
            <Form.Item name="oldPassword" label={t("settings.security.password.form.old_password.label")} rules={[formRule]}>
              <Input.Password autoFocus placeholder={t("settings.security.password.form.old_password.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item name="newPassword" label={t("settings.security.password.form.new_password.label")} rules={[formRule]}>
              <Input.Password placeholder={t("settings.security.password.form.new_password.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item name="confirmPassword" label={t("settings.security.password.form.confirm_password.label")} rules={[formRule]}>
              <Input.Password placeholder={t("settings.security.password.form.confirm_password.placeholder")} onChange={handleInputChange} />
            </Form.Item>

            <Form.Item noStyle>
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
            <Form.Item noStyle>
              <Button onClick={handleEditClick}>{t("settings.security.password.button")}</Button>
            </Form.Item>
          </>
        )}
      </Form>
    </Card>
  );
};

export default SettingsSecurity;
