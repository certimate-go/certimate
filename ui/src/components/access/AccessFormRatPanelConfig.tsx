import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForRatPanel } from "@/domain/access";

type AccessFormRatPanelConfigFieldValues = Nullish<AccessConfigForRatPanel>;

export type AccessFormRatPanelConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormRatPanelConfigFieldValues;
  onValuesChange?: (values: AccessFormRatPanelConfigFieldValues) => void;
};

const initFormModel = (): AccessFormRatPanelConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:8888/",
    accessTokenId: 1,
    accessToken: "",
  };
};

const AccessFormRatPanelConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormRatPanelConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    accessTokenId: z.preprocess((v) => Number(v), z.number().positive(t("access.form.ratpanel_access_token_id.placeholder"))),
    accessToken: z.string().nonempty(t("access.form.ratpanel_access_token.placeholder")),
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
      <Form.Item name="serverUrl" label={t("access.form.ratpanel_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.ratpanel_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name="accessTokenId"
        label={t("access.form.ratpanel_access_token_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ratpanel_access_token_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("access.form.ratpanel_access_token_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="accessToken"
        label={t("access.form.ratpanel_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ratpanel_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ratpanel_access_token.placeholder")} />
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

export default AccessFormRatPanelConfig;
