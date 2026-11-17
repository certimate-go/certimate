import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderZeroSSL = () => {
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
        name={[parentNamePath, "eabKid"]}
        initialValue={initialValues.eabKid}
        dependencies={[parentNamePath, "eabHmacKey"]}
        label={t("access.form.shared_acme_eab_kid.label")}
        rules={[formRule]}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.shared_acme_eab_kid.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "eabHmacKey"]}
        initialValue={initialValues.eabHmacKey}
        dependencies={[parentNamePath, "eabKid"]}
        label={t("access.form.shared_acme_eab_hmac_key.label")}
        rules={[formRule]}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.shared_acme_eab_hmac_key.placeholder")} />
      </Form.Item>

      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.zerossl_eab_auto_access.guide") }}></span>} />
      </Form.Item>

      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.zerossl_eab.guide") }}></span>} />
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

  return z
    .object({
      eabKid: z.string().nullish(),
      eabHmacKey: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      const eabKid = values.eabKid ?? "";
      const eabHmacKey = values.eabHmacKey ?? "";
      if (!!eabKid === !!eabHmacKey) return;

      ctx.addIssue({
        code: "custom",
        message: t("access.form.shared_acme_eab_pair.errmsg"),
        path: [!eabKid ? "eabKid" : "eabHmacKey"],
      });
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderZeroSSL, {
  getInitialValues,
  getSchema,
});

export default _default;
