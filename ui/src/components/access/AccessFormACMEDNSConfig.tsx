import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { type AccessConfigForACMEDNS } from "@/domain/access";

type AccessFormACMEDNSConfigFieldValues = Nullish<AccessConfigForACMEDNS>;

export interface AccessFormACMEDNSConfigProps {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormACMEDNSConfigFieldValues;
  onValuesChange?: (values: AccessFormACMEDNSConfigFieldValues) => void;
}

const initFormModel = (): AccessFormACMEDNSConfigFieldValues => {
  return {
    apiBase: "https://auth.acme-dns.io/",
    storageBaseUrl: "",
    storagePath: "",
  };
};

const AccessFormACMEDNSConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormACMEDNSConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiBase: z.url(t("common.errmsg.url_invalid")),
    storageBaseUrl: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    storagePath: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
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
        name="apiBase"
        label={t("access.form.acmedns_api_base.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmedns_api_base.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.acmedns_api_base.placeholder")} />
      </Form.Item>

      <Form.Item
        name="storageBaseUrl"
        label={t("access.form.acmedns_storage_base_url.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmedns_storage_base_url.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.acmedns_storage_base_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="storagePath"
        label={t("access.form.acmedns_storage_path.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmedns_storage_path.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.acmedns_storage_path.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormACMEDNSConfig;
