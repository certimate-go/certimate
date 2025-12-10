import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import TextFileInput from "@/components/TextFileInput";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFieldsProviderACMEDNS = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item
        name={[parentNamePath, "serverUrl"]}
        initialValue={initialValues.serverUrl}
        label={t("access.form.acmedns_server_url.label")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.acmedns_server_url.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "credentials"]}
        initialValue={initialValues.credentials}
        label={t("access.form.acmedns_credentials.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmedns_credentials.tooltip") }}></span>}
      >
        <TextFileInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("access.form.acmedns_credentials.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    serverUrl: "https://auth.acme-dns.io/",
    credentials: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    serverUrl: z.url(t("common.errmsg.url_invalid")),
    credentials: z.string().refine((v) => {
      if (!v) return false;

      try {
        const obj = JSON.parse(v);
        return typeof obj === "object" && !Array.isArray(obj);
      } catch {
        return false;
      }
    }, t("access.form.acmedns_credentials.errmsg.json_invalid")),
  });
};

const _default = Object.assign(AccessConfigFieldsProviderACMEDNS, {
  getInitialValues,
  getSchema,
});

export default _default;
