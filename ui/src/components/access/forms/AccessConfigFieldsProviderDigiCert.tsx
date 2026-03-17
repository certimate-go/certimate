import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderDigiCert = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item name={[parentNamePath, "eabKid"]} initialValue={initialValues.eabKid} label={t("access.form.shared_acme_eab_kid.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.shared_acme_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "eabHmacKey"]}
        initialValue={initialValues.eabHmacKey}
        label={t("access.form.shared_acme_eab_hmac_key.label")}
        rules={[formRule]}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.shared_acme_eab_hmac_key.placeholder")} />
      </Form.Item>

      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.digicert_eab.guide") }}></span>} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    eabKid: "",
    eabHmacKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    eabKid: z.string().nonempty(t("access.form.shared_acme_eab_kid.placeholder")),
    eabHmacKey: z.string().nonempty(t("access.form.shared_acme_eab_hmac_key.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderDigiCert, {
  getInitialValues,
  getSchema,
});

export default _default;
