import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForLarkBot } from "@/domain/access";

type AccessFormLarkBotConfigFieldValues = Nullish<AccessConfigForLarkBot>;

export type AccessFormLarkBotConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormLarkBotConfigFieldValues;
  onValuesChange?: (values: AccessFormLarkBotConfigFieldValues) => void;
};

const initFormModel = (): AccessFormLarkBotConfigFieldValues => {
  return {
    webhookUrl: "",
  };
};

const AccessFormLarkBotConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormLarkBotConfigProps) => {
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
        label={t("access.form.larkbot_webhook_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.larkbot_webhook_url.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.larkbot_webhook_url.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormLarkBotConfig;
