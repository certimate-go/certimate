import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderCloudflare = () => {
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
        name={[parentNamePath, "dnsApiToken"]}
        initialValue={initialValues.dnsApiToken}
        label={t("access.form.cloudflare_dns_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudflare_dns_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cloudflare_dns_api_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "zoneApiToken"]}
        initialValue={initialValues.zoneApiToken}
        label={t("access.form.cloudflare_zone_api_token.label")}
        extra={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudflare_zone_api_token.help") }}></span>}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cloudflare_zone_api_token.tooltip") }}></span>}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.cloudflare_zone_api_token.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    dnsApiToken: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    dnsApiToken: z.string().nonempty(),
    zoneApiToken: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderCloudflare, {
  getInitialValues,
  getSchema,
});

export default _default;
