import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForClouDNS } from "@/domain/access";

type AccessFormClouDNSConfigFieldValues = Nullish<AccessConfigForClouDNS>;

export type AccessFormClouDNSConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormClouDNSConfigFieldValues;
  onValuesChange?: (values: AccessFormClouDNSConfigFieldValues) => void;
};

const initFormModel = (): AccessFormClouDNSConfigFieldValues => {
  return {
    authId: "",
    authPassword: "",
  };
};

const AccessFormClouDNSConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormClouDNSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    authId: z
      .string()
      .min(1, t("access.form.cloudns_auth_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    authPassword: z
      .string()
      .min(1, t("access.form.cloudns_auth_password.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
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
        name="authId"
        label={t("access.form.cloudns_auth_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudns_auth_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.cloudns_auth_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="authPassword"
        label={t("access.form.cloudns_auth_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudns_auth_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cloudns_auth_password.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormClouDNSConfig;
