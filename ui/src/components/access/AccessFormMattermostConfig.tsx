import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForMattermost } from "@/domain/access";

type AccessFormMattermostConfigFieldValues = Nullish<AccessConfigForMattermost>;

export type AccessFormMattermostConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormMattermostConfigFieldValues;
  onValuesChange?: (values: AccessFormMattermostConfigFieldValues) => void;
};

const initFormModel = (): AccessFormMattermostConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:8065/",
    username: "",
    password: "",
  };
};

const AccessFormMattermostConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormMattermostConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    username: z.string().nonempty(t("access.form.mattermost_username.placeholder")),
    password: z.string().nonempty(t("access.form.mattermost_password.placeholder")),
    defaultChannelId: z.string().nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
    onValuesChange?.(values);
  };

  return (
    <Form
      form={formInst}
      disabled={disabled}
      initialValues={initialValues ?? initFormModel()}
      layout="vertical"
      name={formName}
      onValuesChange={handleFormChange}
    >
      <Form.Item
        name="serverUrl"
        label={t("access.form.mattermost_server_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.mattermost_server_url.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.mattermost_server_url.placeholder")} />
      </Form.Item>

      <Form.Item name="username" label={t("access.form.mattermost_username.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.mattermost_username.placeholder")} />
      </Form.Item>

      <Form.Item name="password" label={t("access.form.mattermost_password.label")} rules={[formRule]}>
        <Input.Password placeholder={t("access.form.mattermost_password.placeholder")} />
      </Form.Item>

      <Form.Item
        name="defaultChannelId"
        label={t("access.form.mattermost_default_channel_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.mattermost_default_channel_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.mattermost_default_channel_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormMattermostConfig;
