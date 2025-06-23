import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForSlackBot } from "@/domain/access";

type AccessFormSlackBotConfigFieldValues = Nullish<AccessConfigForSlackBot>;

export type AccessFormSlackBotConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormSlackBotConfigFieldValues;
  onValuesChange?: (values: AccessFormSlackBotConfigFieldValues) => void;
};

const initFormModel = (): AccessFormSlackBotConfigFieldValues => {
  return {
    botToken: "",
  };
};

const AccessFormSlackBotConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormSlackBotConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    botToken: z
      .string(t("access.form.slackbot_token.placeholder"))
      .min(1, t("access.form.slackbot_token.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
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
        name="botToken"
        label={t("access.form.slackbot_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.slackbot_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.slackbot_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name="defaultChannelId"
        label={t("access.form.slackbot_default_channel_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.slackbot_default_channel_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.slackbot_default_channel_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormSlackBotConfig;
