import { getI18n, useTranslation } from "react-i18next";
import { Checkbox, Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";
import { isJsonObject } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderLarkBot = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance<z.infer<typeof formSchema>>();
  const initialValues = getInitialValues();

  const fieldUseCustomPayload = Form.useWatch([parentNamePath, "useCustomPayload"], formInst);

  const handleCustomPayloadChecked = (checked: boolean) => {
    formInst.setFieldValue([parentNamePath, "useCustomPayload"], checked);
    if (checked) {
      formInst.setFieldValue([parentNamePath, "customPayload"], commonPayloadString);
    } else {
      formInst.setFieldValue([parentNamePath, "customPayload"], void 0);
    }
  };

  const handleCustomPayloadBlur = () => {
    const value = formInst.getFieldValue([parentNamePath, "customPayload"]);
    try {
      const json = JSON.stringify(JSON.parse(value), null, 2);
      formInst.setFieldValue([parentNamePath, "customPayload"], json);
    } catch {
      return;
    }
  };

  return (
    <>
      <Form.Item
        name={[parentNamePath, "webhookUrl"]}
        initialValue={initialValues.webhookUrl}
        label={t("access.form.larkbot_webhook_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.larkbot_webhook_url.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.larkbot_webhook_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secret"]}
        initialValue={initialValues.secret}
        label={t("access.form.larkbot_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.larkbot_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.larkbot_secret.placeholder")} />
      </Form.Item>

      <Form.Item label={t("access.form.larkbot_custom_payload.label")}>
        <Form.Item name={[parentNamePath, "useCustomPayload"]} noStyle>
          <Checkbox checked={!!fieldUseCustomPayload} onChange={(e) => handleCustomPayloadChecked(e.target.checked)}>
            {t("access.form.larkbot_custom_payload.checkbox")}
          </Checkbox>
        </Form.Item>
        <Form.Item
          name={[parentNamePath, "customPayload"]}
          hidden={!fieldUseCustomPayload}
          initialValue={initialValues.customPayload}
          noStyle
          rules={[formRule]}
        >
          <CodeTextInput
            className="mt-2"
            lineWrapping={false}
            height="auto"
            minHeight="64px"
            maxHeight="256px"
            language="json"
            placeholder={t("access.form.larkbot_custom_payload.placeholder")}
            onBlur={handleCustomPayloadBlur}
          />
        </Form.Item>
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    webhookUrl: "",
    secret: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      webhookUrl: z.url(t("common.errmsg.url_invalid")),
      secret: z.string().nullish(),
      useCustomPayload: z.boolean().nullish(),
      customPayload: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.useCustomPayload) {
        if (!isJsonObject(values.customPayload!)) {
          ctx.addIssue({
            code: "custom",
            message: t("common.errmsg.json_invalid"),
            path: ["customPayload"],
          });
        }
      }
    });
};

const commonPayloadString = JSON.stringify(
  {
    msg_type: "text",
    content: {
      text: "${CERTIMATE_NOTIFIER_SUBJECT}\n\n${CERTIMATE_NOTIFIER_MESSAGE}",
    },
  },
  null,
  2
);

const _default = Object.assign(AccessConfigFormFieldsProviderLarkBot, {
  getInitialValues,
  getSchema,
});

export default _default;
