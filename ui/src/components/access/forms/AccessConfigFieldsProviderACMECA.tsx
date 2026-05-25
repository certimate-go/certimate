import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { core, z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderACMECA = () => {
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
        name={[parentNamePath, "endpoint"]}
        initialValue={initialValues.endpoint}
        label={t("access.form.acmeca_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmeca_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.acmeca_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "eabKid"]} initialValue={initialValues.eabKid} label={t("access.form.acmeca_eab_kid.label")} rules={[formRule]}>
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.acmeca_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "eabHmacKey"]}
        initialValue={initialValues.eabHmacKey}
        label={t("access.form.acmeca_eab_hmac_key.label")}
        rules={[formRule]}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.acmeca_eab_hmac_key.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    endpoint: "https://example.com/acme/directory",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    endpoint: z.url({ protocol: core.regexes.httpProtocol }),
    eabKid: z.string().nullish(),
    eabHmacKey: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderACMECA, {
  getInitialValues,
  getSchema,
});

export default _default;
