import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizNotifyNodeConfigFieldsProviderTelegramBot = () => {
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
        name={[parentNamePath, "chatId"]}
        initialValue={initialValues.chatId}
        label={t("workflow_node.notify.form.telegrambot_chat_id.label")}
        extra={t("workflow_node.notify.form.telegrambot_chat_id.help")}
        rules={[formRule]}
      >
        <Input allowClear placeholder={t("workflow_node.notify.form.telegrambot_chat_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {};
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    chatId: z.union([z.string(), z.int()]).nullish(),
  });
};

const _default = Object.assign(BizNotifyNodeConfigFieldsProviderTelegramBot, {
  getInitialValues,
  getSchema,
});

export default _default;
