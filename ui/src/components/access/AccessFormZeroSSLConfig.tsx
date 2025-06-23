import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForZeroSSL } from "@/domain/access";

type AccessFormZeroSSLConfigFieldValues = Nullish<AccessConfigForZeroSSL>;

export type AccessFormZeroSSLConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormZeroSSLConfigFieldValues;
  onValuesChange?: (values: AccessFormZeroSSLConfigFieldValues) => void;
};

const initFormModel = (): AccessFormZeroSSLConfigFieldValues => {
  return {
    eabKid: "",
    eabHmacKey: "",
  };
};

const AccessFormZeroSSLConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormZeroSSLConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    eabKid: z
      .string()
      .min(1, t("access.form.zerossl_eab_kid.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    eabHmacKey: z
      .string()
      .min(1, t("access.form.zerossl_eab_hmac_key.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
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
        name="eabKid"
        label={t("access.form.zerossl_eab_kid.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.zerossl_eab_kid.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.zerossl_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item
        name="eabHmacKey"
        label={t("access.form.zerossl_eab_hmac_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.zerossl_eab_hmac_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.zerossl_eab_hmac_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormZeroSSLConfig;
