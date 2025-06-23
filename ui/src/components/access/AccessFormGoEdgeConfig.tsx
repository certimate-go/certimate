import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Radio, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForGoEdge } from "@/domain/access";

type AccessFormGoEdgeConfigFieldValues = Nullish<AccessConfigForGoEdge>;

export type AccessFormGoEdgeConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormGoEdgeConfigFieldValues;
  onValuesChange?: (values: AccessFormGoEdgeConfigFieldValues) => void;
};

const initFormModel = (): AccessFormGoEdgeConfigFieldValues => {
  return {
    serverUrl: "http://<your-host-addr>:7788/",
    apiRole: "user",
    accessKeyId: "",
    accessKey: "",
  };
};

const AccessFormGoEdgeConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormGoEdgeConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    role: z.literal(["user", "admin"], t("access.form.goedge_api_role.placeholder")),
    accessKeyId: z.string().nonempty(t("access.form.goedge_access_key_id.placeholder")),
    accessKey: z.string().nonempty(t("access.form.goedge_access_key.placeholder")),
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
      <Form.Item name="serverUrl" label={t("access.form.goedge_server_url.label")} rules={[formRule]}>
        <Input placeholder={t("access.form.goedge_server_url.placeholder")} />
      </Form.Item>

      <Form.Item name="apiRole" label={t("access.form.goedge_api_role.label")} rules={[formRule]}>
        <Radio.Group options={["user", "admin"].map((s) => ({ label: t(`access.form.goedge_api_role.option.${s}.label`), value: s }))} />
      </Form.Item>

      <Form.Item
        name="accessKeyId"
        label={t("access.form.goedge_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.goedge_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.goedge_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="accessKey"
        label={t("access.form.goedge_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.goedge_access_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.goedge_access_key.placeholder")} />
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

export default AccessFormGoEdgeConfig;
