import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isEmail } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const MESSAGE_FORMAT_PLAIN = "plain" as const;
const MESSAGE_FORMAT_HTML = "html" as const;

const BizNotifyNodeConfigFieldsProviderEmail = () => {
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
        name={[parentNamePath, "format"]}
        initialValue={initialValues.format}
        label={t("workflow_node.notify.form.email_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[MESSAGE_FORMAT_PLAIN, MESSAGE_FORMAT_HTML].map((s) => ({
            key: s,
            label: t(`workflow_node.notify.form.email_format.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.notify.form.email_format.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "receiverAddress"]}
        initialValue={initialValues.receiverAddress}
        label={t("workflow_node.notify.form.email_receiver_address.label")}
        extra={t("workflow_node.notify.form.email_receiver_address.help")}
        rules={[formRule]}
      >
        <Input type="email" allowClear placeholder={t("workflow_node.notify.form.email_receiver_address.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    format: MESSAGE_FORMAT_PLAIN,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    format: z.enum([MESSAGE_FORMAT_PLAIN, MESSAGE_FORMAT_HTML]).nullish(),
    receiverAddress: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return isEmail(v);
      }, t("common.errmsg.email_invalid")),
  });
};

const _default = Object.assign(BizNotifyNodeConfigFieldsProviderEmail, {
  getInitialValues,
  getSchema,
});

export default _default;
