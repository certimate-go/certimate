import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForACMECA } from "@/domain/access";

type AccessFormACMECAConfigFieldValues = Nullish<AccessConfigForACMECA>;

export type AccessFormACMECAConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormACMECAConfigFieldValues;
  onValuesChange?: (values: AccessFormACMECAConfigFieldValues) => void;
};

const initFormModel = (): AccessFormACMECAConfigFieldValues => {
  return {
    endpoint: "https://example.com/acme/directory",
  };
};

const AccessFormACMECAConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormACMECAConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.url(t("common.errmsg.url_invalid")),
    eabKid: z.string().nullish(),
    eabHmacKey: z.string().nullish(),
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
        name="endpoint"
        label={t("access.form.acmeca_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.acmeca_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="eabKid"
        label={t("access.form.acmeca_eab_kid.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_eab_kid.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.acmeca_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item
        name="eabHmacKey"
        label={t("access.form.acmeca_eab_hmac_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_eab_hmac_key.tooltip") }}></span>}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.acmeca_eab_hmac_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormACMECAConfig;
