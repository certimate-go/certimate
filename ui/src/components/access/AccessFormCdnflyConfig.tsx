import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForCdnfly } from "@/domain/access";

type AccessFormCdnflyConfigFieldValues = Nullish<AccessConfigForCdnfly>;

export type AccessFormCdnflyConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormCdnflyConfigFieldValues;
  onValuesChange?: (values: AccessFormCdnflyConfigFieldValues) => void;
};

const initFormModel = (): AccessFormCdnflyConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:88/",
    apiKey: "",
    apiSecret: "",
  };
};

const AccessFormCdnflyConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormCdnflyConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    apiKey: z
      .string()
      .min(1, t("access.form.cdnfly_api_key.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiSecret: z
      .string()
      .min(1, t("access.form.cdnfly_api_secret.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    allowInsecureConnections: z.boolean().nullish(),
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
      <Form.Item name="serverUrl" label={t("access.form.cdnfly_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.cdnfly_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiKey"
        label={t("access.form.cdnfly_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cdnfly_api_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.cdnfly_api_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiSecret"
        label={t("access.form.cdnfly_api_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cdnfly_api_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cdnfly_api_secret.placeholder")} />
      </Form.Item>

      <Form.Item name="allowInsecureConnections" label={t("access.form.common_allow_insecure_conns.label")} rules={[formRule]}>
        <Switch
          checkedChildren={t("access.form.common_allow_insecure_conns.switch.on")}
          unCheckedChildren={t("access.form.common_allow_insecure_conns.switch.off")}
        />
      </Form.Item>
    </Form>
  );
};

export default AccessFormCdnflyConfig;
