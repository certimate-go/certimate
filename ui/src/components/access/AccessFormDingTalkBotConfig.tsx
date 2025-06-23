import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForDingTalkBot } from "@/domain/access";

type AccessFormDingTalkBotConfigFieldValues = Nullish<AccessConfigForDingTalkBot>;

export type AccessFormDingTalkBotConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormDingTalkBotConfigFieldValues;
  onValuesChange?: (values: AccessFormDingTalkBotConfigFieldValues) => void;
};

const initFormModel = (): AccessFormDingTalkBotConfigFieldValues => {
  return {
    webhookUrl: "",
    secret: "",
  };
};

const AccessFormDingTalkBotConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormDingTalkBotConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    webhookUrl: z.url(t("common.errmsg.url_invalid")),
    secret: z.string().nonempty(t("access.form.dingtalkbot_secret.placeholder")),
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
        name="webhookUrl"
        label={t("access.form.dingtalkbot_webhook_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dingtalkbot_webhook_url.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.dingtalkbot_webhook_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secret"
        label={t("access.form.dingtalkbot_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.dingtalkbot_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.dingtalkbot_secret.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormDingTalkBotConfig;
