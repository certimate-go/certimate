import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import FileTextInput from "@/components/FileTextInput";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderRegru = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item name={[parentNamePath, "username"]} initialValue={initialValues.username} label={t("access.form.regru_username.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.regru_username.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "password"]} initialValue={initialValues.password} label={t("access.form.regru_password.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.regru_password.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "mtlsCertificate"]}
        initialValue={initialValues.mtlsCertificate}
        label={t("access.form.regru_mtls_certificate.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.regru_mtls_certificate.tooltip") }}></span>}
      >
        <FileTextInput autoSize={{ minRows: 1, maxRows: 5 }} placeholder={t("access.form.regru_mtls_certificate.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "mtlsPrivateKey"]}
        initialValue={initialValues.mtlsPrivateKey}
        label={t("access.form.regru_mtls_private_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.regru_mtls_private_key.tooltip") }}></span>}
      >
        <FileTextInput autoSize={{ minRows: 1, maxRows: 5 }} placeholder={t("access.form.regru_mtls_private_key.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    username: "",
    password: "",
    mtlsCertificate: "",
    mtlsPrivateKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      username: z.string().nonempty(),
      password: z.string().nonempty(),
      mtlsCertificate: z.string().nullish(),
      mtlsPrivateKey: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      const scMtlsCertificate = z.string().nonempty();
      const spMtlsCertificate = scMtlsCertificate.safeParse(values.mtlsCertificate);
      const scMtlsPrivateKey = z.string().nonempty();
      const spMtlsPrivateKey = scMtlsPrivateKey.safeParse(values.mtlsPrivateKey);
      if (!spMtlsCertificate.success && spMtlsPrivateKey.success) {
        ctx.addIssue({
          code: "custom",
          message: t("access.form.regru_mtls_certificate.errmsg.invalid"),
          path: ["mtlsCertificate"],
        });
      } else if (spMtlsCertificate.success && !spMtlsPrivateKey.success) {
        ctx.addIssue({
          code: "custom",
          message: t("access.form.regru_mtls_private_key.errmsg.invalid"),
          path: ["mtlsPrivateKey"],
        });
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderRegru, {
  getInitialValues,
  getSchema,
});

export default _default;
