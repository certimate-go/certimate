import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderMattermost = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item
        name={[parentNamePath, "serverUrl"]}
        initialValue={initialValues.serverUrl}
        label={t("access.form.mattermost_server_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.mattermost_server_url.tooltip") }}></span>}
      >
        <Input type="url" placeholder={t("access.form.mattermost_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "username"]}
        initialValue={initialValues.username}
        label={t("access.form.mattermost_username.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("access.form.mattermost_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "password"]}
        initialValue={initialValues.password}
        label={t("access.form.mattermost_password.label")}
        rules={[formRule]}
      >
        <Input.Password placeholder={t("access.form.mattermost_password.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "channelId"]}
        initialValue={initialValues.channelId}
        label={t("access.form.mattermost_channel_id.label")}
        extra={t("access.form.mattermost_channel_id.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.mattermost_channel_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.mattermost_channel_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    serverUrl: "http://<your-host-addr>:8065/",
    username: "",
    password: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    username: z.string().nonempty(t("access.form.mattermost_username.placeholder")),
    password: z.string().nonempty(t("access.form.mattermost_password.placeholder")),
    channelId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderMattermost, {
  getInitialValues,
  getSchema,
});

export default _default;
