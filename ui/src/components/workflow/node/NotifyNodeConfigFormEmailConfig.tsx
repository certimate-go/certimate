import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validEmailAddress } from "@/utils/validators";

type NotifyNodeConfigFormEmailConfigFieldValues = Nullish<{
  senderAddress?: string;
  receiverAddress?: string;
}>;

export type NotifyNodeConfigFormEmailConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: NotifyNodeConfigFormEmailConfigFieldValues;
  onValuesChange?: (values: NotifyNodeConfigFormEmailConfigFieldValues) => void;
};

const initFormModel = (): NotifyNodeConfigFormEmailConfigFieldValues => {
  return {};
};

const NotifyNodeConfigFormEmailConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: NotifyNodeConfigFormEmailConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    senderAddress: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return validEmailAddress(v);
      }, t("common.errmsg.email_invalid")),
    senderName: z.string().nullish(),
    receiverAddress: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return validEmailAddress(v);
      }, t("common.errmsg.email_invalid")),
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
        name="senderAddress"
        label={t("workflow_node.notify.form.email_sender_address.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.email_sender_address.tooltip") }}></span>}
      >
        <Input type="email" allowClear placeholder={t("workflow_node.notify.form.email_sender_address.placeholder")} />
      </Form.Item>

      <Form.Item
        name="senderName"
        label={t("workflow_node.notify.form.email_sender_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.email_sender_name.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.notify.form.email_sender_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name="receiverAddress"
        label={t("workflow_node.notify.form.email_receiver_address.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.notify.form.email_receiver_address.tooltip") }}></span>}
      >
        <Input type="email" allowClear placeholder={t("workflow_node.notify.form.email_receiver_address.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default NotifyNodeConfigFormEmailConfig;
