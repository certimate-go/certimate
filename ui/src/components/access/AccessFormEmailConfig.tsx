import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, InputNumber, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForEmail } from "@/domain/access";
import { validEmailAddress, validPortNumber } from "@/utils/validators";

type AccessFormEmailConfigFieldValues = Nullish<AccessConfigForEmail>;

export type AccessFormEmailConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormEmailConfigFieldValues;
  onValuesChange?: (values: AccessFormEmailConfigFieldValues) => void;
};

const initFormModel = (): AccessFormEmailConfigFieldValues => {
  return {
    smtpHost: "",
    smtpPort: 465,
    smtpTls: true,
    username: "",
    password: "",
  };
};

const AccessFormEmailConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormEmailConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    smtpHost: z
      .string()
      .min(1, t("access.form.email_smtp_host.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    smtpPort: z.preprocess(
      (v) => Number(v),
      z.number().refine((v) => validPortNumber(v), t("common.errmsg.port_invalid"))
    ),
    smtpTls: z.boolean().nullish(),
    username: z
      .string()
      .min(1, t("access.form.email_username.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    password: z
      .string()
      .min(1, t("access.form.email_password.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    defaultSenderAddress: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return validEmailAddress(v);
      }, t("common.errmsg.email_invalid")),
    defaultSenderName: z.string().nullish(),
    defaultReceiverAddress: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return validEmailAddress(v);
      }, t("common.errmsg.email_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const handleTlsSwitchChange = (checked: boolean) => {
    const oldPort = formInst.getFieldValue("smtpPort");
    const newPort = checked && (oldPort == null || oldPort === 25) ? 465 : !checked && (oldPort == null || oldPort === 465) ? 25 : oldPort;
    if (newPort !== oldPort) {
      formInst.setFieldValue("smtpPort", newPort);
    }
  };

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
      <div className="flex space-x-2">
        <div className="w-3/5">
          <Form.Item name="smtpHost" label={t("access.form.email_smtp_host.label")} rules={[formRule]}>
            <Input placeholder={t("access.form.email_smtp_host.placeholder")} />
          </Form.Item>
        </div>

        <div className="w-2/5">
          <Form.Item name="smtpPort" label={t("access.form.email_smtp_port.label")} rules={[formRule]}>
            <InputNumber style={{ width: "100%" }} placeholder={t("access.form.email_smtp_port.placeholder")} min={1} max={65535} />
          </Form.Item>
        </div>
      </div>

      <Form.Item name="smtpTls" label={t("access.form.email_smtp_tls.label")} rules={[formRule]}>
        <Switch onChange={handleTlsSwitchChange} />
      </Form.Item>

      <Form.Item name="username" label={t("access.form.email_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.email_username.placeholder")} />
      </Form.Item>

      <Form.Item name="password" label={t("access.form.email_password.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.email_password.placeholder")} />
      </Form.Item>

      <Form.Item name="defaultSenderAddress" label={t("access.form.email_default_sender_address.label")} rules={[formRule]}>
        <Input type="email" allowClear placeholder={t("access.form.email_default_sender_address.placeholder")} />
      </Form.Item>

      <Form.Item name="defaultSenderName" label={t("access.form.email_default_sender_name.label")} rules={[formRule]}>
        <Input allowClear placeholder={t("access.form.email_default_sender_name.placeholder")} />
      </Form.Item>

      <Form.Item name="defaultReceiverAddress" label={t("access.form.email_default_receiver_address.label")} rules={[formRule]}>
        <Input type="email" allowClear placeholder={t("access.form.email_default_receiver_address.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormEmailConfig;
