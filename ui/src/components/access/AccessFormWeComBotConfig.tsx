import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForWeComBot } from "@/domain/access";

type AccessFormWeComBotConfigFieldValues = Nullish<AccessConfigForWeComBot>;

export type AccessFormWeComBotConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormWeComBotConfigFieldValues;
  onValuesChange?: (values: AccessFormWeComBotConfigFieldValues) => void;
};

const initFormModel = (): AccessFormWeComBotConfigFieldValues => {
  return {
    webhookUrl: "",
  };
};

const AccessFormWeComBotConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormWeComBotConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    webhookUrl: z.url(t("common.errmsg.url_invalid")),
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
        label={t("access.form.wecombot_webhook_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.wecombot_webhook_url.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.wecombot_webhook_url.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormWeComBotConfig;
