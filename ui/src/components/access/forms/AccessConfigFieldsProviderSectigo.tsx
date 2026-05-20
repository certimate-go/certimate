import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderSectigo = () => {
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
        name={[parentNamePath, "validationType"]}
        initialValue={initialValues.validationType}
        label={t("access.form.sectigo_validation_type.label")}
        rules={[formRule]}
      >
        <Select
          options={["dv", "ov", "ev"].map((s) => ({
            key: s,
            label: t(`access.form.sectigo_validation_type.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("access.form.sectigo_validation_type.placeholder")}
        />
      </Form.Item>

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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.sectigo_eab.guide") }}></span>} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    validationType: "dv",
    eabKid: "",
    eabHmacKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    validationType: z.string().nonempty(),
    eabKid: z.string().nonempty(),
    eabHmacKey: z.string().nonempty(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderSectigo, {
  getInitialValues,
  getSchema,
});

export default _default;
