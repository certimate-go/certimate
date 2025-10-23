import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, InputNumber, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { validDomainName, validIPv4Address, validIPv6Address, validPortNumber } from "@/utils/validators";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderRFC2136 = () => {
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
          <Form.Item name={[parentNamePath, "host"]} initialValue={initialValues.host} label={t("access.form.rfc2136_host.label")} rules={[formRule]}>
            <Input placeholder={t("access.form.rfc2136_host.placeholder")} />
          </Form.Item>
        </div>

        <div className="w-1/3">
          <Form.Item name={[parentNamePath, "port"]} initialValue={initialValues.port} label={t("access.form.rfc2136_port.label")} rules={[formRule]}>
            <InputNumber style={{ width: "100%" }} min={1} max={65535} placeholder={t("access.form.rfc2136_port.placeholder")} />
          </Form.Item>
        </div>
      </div>

      <Form.Item
        name={[parentNamePath, "tsigAlgorithm"]}
        initialValue={initialValues.tsigAlgorithm}
        label={t("access.form.rfc2136_tsig_algorithm.label")}
        rules={[formRule]}
      >
        <Select
          options={[
            { label: "HMAC-SHA-1", value: "hmac-sha1." },
            { label: "HMAC-SHA-224", value: "hmac-sha224." },
            { label: "HMAC-SHA-256", value: "hmac-sha256." },
            { label: "HMAC-SHA-384", value: "hmac-sha384." },
            { label: "HMAC-SHA-512", value: "hmac-sha512." },
          ]}
          placeholder={t("access.form.rfc2136_tsig_algorithm.placeholder")}
        />
      </Form.Item>

      <Form.Item name={[parentNamePath, "tsigKey"]} initialValue={initialValues.tsigKey} label={t("access.form.rfc2136_tsig_key.label")} rules={[formRule]}>
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.rfc2136_tsig_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "tsigSecret"]}
        initialValue={initialValues.tsigSecret}
        label={t("access.form.rfc2136_tsig_secret.label")}
        rules={[formRule]}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.rfc2136_tsig_secret.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    host: "127.0.0.1",
    port: 53,
    tsigAlgorithm: "hmac-sha1.",
    tsigKey: "",
    tsigSecret: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    host: z.string().refine((v) => validDomainName(v) || validIPv4Address(v) || validIPv6Address(v), t("common.errmsg.host_invalid")),
    port: z.preprocess(
      (v) => Number(v),
      z
        .number()
        .int(t("access.form.rfc2136_port.placeholder"))
        .refine((v) => validPortNumber(v), t("common.errmsg.port_invalid"))
    ),
    tsigAlgorithm: z.string().nonempty(t("access.form.rfc2136_tsig_algorithm.placeholder")),
    tsigKey: z.string().nullish(),
    tsigSecret: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderRFC2136, {
  getInitialValues,
  getSchema,
});

export default _default;
