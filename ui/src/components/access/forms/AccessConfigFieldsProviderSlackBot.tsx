import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderSlackBot = () => {
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
        name={[parentNamePath, "botToken"]}
        initialValue={initialValues.botToken}
        label={t("access.form.slackbot_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.slackbot_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.slackbot_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "channelId"]}
        initialValue={initialValues.channelId}
        label={t("access.form.slackbot_channel_id.label")}
        extra={t("access.form.slackbot_channel_id.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.slackbot_channel_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.slackbot_channel_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    botToken: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    botToken: z.string().nonempty(),
    channelId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderSlackBot, {
  getInitialValues,
  getSchema,
});

export default _default;
