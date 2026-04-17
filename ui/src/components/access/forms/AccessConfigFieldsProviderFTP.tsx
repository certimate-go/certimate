import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, InputNumber } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isHostname, isPortNumber } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderFTP = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <div className="flex space-x-2">
        <div className="w-2/3">
          <Form.Item name={[parentNamePath, "host"]} initialValue={initialValues.host} label={t("access.form.ftp_host.label")} rules={[formRule]}>
            <Input placeholder={t("access.form.ftp_host.placeholder")} />
          </Form.Item>
        </div>

        <div className="w-1/3">
          <Form.Item name={[parentNamePath, "port"]} initialValue={initialValues.port} label={t("access.form.ftp_port.label")} rules={[formRule]}>
            <InputNumber style={{ width: "100%" }} min={1} max={65535} placeholder={t("access.form.ftp_port.placeholder")} />
          </Form.Item>
        </div>
      </div>

      <Form.Item name={[parentNamePath, "username"]} initialValue={initialValues.username} label={t("access.form.ftp_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.ftp_username.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "password"]} initialValue={initialValues.password} label={t("access.form.ftp_password.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ftp_password.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    host: "127.0.0.1",
    port: 21,
    username: "",
    password: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    host: z.string().refine((v) => isHostname(v), t("common.errmsg.host_invalid")),
    port: z.coerce.number().refine((v) => isPortNumber(v), t("common.errmsg.port_invalid")),
    username: z.string().nonempty(t("access.form.ftp_username.placeholder")),
    password: z.string().nonempty(t("access.form.ftp_password.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderFTP, {
  getInitialValues,
  getSchema,
});

export default _default;
