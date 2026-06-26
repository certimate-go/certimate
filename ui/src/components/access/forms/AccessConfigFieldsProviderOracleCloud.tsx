import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import FileTextInput from "@/components/FileTextInput";
import Show from "@/components/Show";
import { validatePEMPrivateKey } from "@/utils/x509";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_METHOD_APIKEY = "apikey" as const;
const AUTH_METHOD_INSTANCEPRINCIPAL = "instanceprincipal" as const;
const AUTH_METHOD_RESOURCEPRINCIPAL = "resourceprincipal" as const;

const AccessConfigFieldsProviderOracleCloud = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance<z.infer<typeof formSchema>>();
  const initialValues = getInitialValues();

  const fieldAuthMethod = Form.useWatch<string>([parentNamePath, "authMethod"], formInst);

  return (
    <>
      <Form.Item
        name={[parentNamePath, "authMethod"]}
        initialValue={initialValues.authMethod}
        label={t("access.form.oraclecloud_auth_method.label")}
        rules={[formRule]}
      >
        <Radio.Group block>
          <Radio.Button value={AUTH_METHOD_APIKEY}>{t("access.form.oraclecloud_auth_method.option.apikey.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_INSTANCEPRINCIPAL}>{t("access.form.oraclecloud_auth_method.option.instanceprincipal.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_RESOURCEPRINCIPAL}>{t("access.form.oraclecloud_auth_method.option.resourceprincipal.label")}</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_APIKEY}>
        <Form.Item
          name={[parentNamePath, "privateKey"]}
          initialValue={initialValues.privateKey}
          label={t("access.form.oraclecloud_private_key.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.oraclecloud_private_key.tooltip") }}></span>}
        >
          <FileTextInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("access.form.oraclecloud_private_key.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "privateKeyPassphrase"]}
          initialValue={initialValues.privateKeyPassphrase}
          label={t("access.form.oraclecloud_private_key_passphrase.label")}
          rules={[formRule]}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.oraclecloud_private_key_passphrase.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "publicKeyFingerprint"]}
          initialValue={initialValues.publicKeyFingerprint}
          label={t("access.form.oraclecloud_public_key_fingerprint.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("access.form.oraclecloud_public_key_fingerprint.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "tenancyOcid"]}
          initialValue={initialValues.tenancyOcid}
          label={t("access.form.oraclecloud_tenancy_ocid.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("access.form.oraclecloud_tenancy_ocid.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "userOcid"]}
          initialValue={initialValues.userOcid}
          label={t("access.form.oraclecloud_user_ocid.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("access.form.oraclecloud_user_ocid.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    authMethod: AUTH_METHOD_APIKEY,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      authMethod: z.enum([AUTH_METHOD_APIKEY, AUTH_METHOD_INSTANCEPRINCIPAL, AUTH_METHOD_RESOURCEPRINCIPAL]),
      privateKey: z.string().nullish(),
      privateKeyPassphrase: z.string().nullish(),
      publicKeyFingerprint: z.string().nullish(),
      tenancyOcid: z.string().nullish(),
      userOcid: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.authMethod) {
        case AUTH_METHOD_APIKEY:
          {
            const scPrivateKey = z.string().refine((v) => validatePEMPrivateKey(v));
            const spPrivateKey = scPrivateKey.safeParse(values.privateKey);
            if (!spPrivateKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPrivateKey.error).errors.join(),
                path: ["privateKey"],
              });
            }

            const scPublicKeyFingerprint = z.string().regex(/^[0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){15}$/);
            const spPublicKeyFingerprint = scPublicKeyFingerprint.safeParse(values.publicKeyFingerprint);
            if (!spPublicKeyFingerprint.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPublicKeyFingerprint.error).errors.join(),
                path: ["publicKeyFingerprint"],
              });
            }

            const scTenancyOcid = z.string().regex(/^ocid\d\..{1,}$/);
            const spTenancyOcid = scTenancyOcid.safeParse(values.tenancyOcid);
            if (!spTenancyOcid.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spTenancyOcid.error).errors.join(),
                path: ["tenancyOcid"],
              });
            }

            const scUserOcid = z.string().regex(/^ocid\d\..{1,}$/);
            const spUserOcid = scUserOcid.safeParse(values.userOcid);
            if (!spUserOcid.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spUserOcid.error).errors.join(),
                path: ["userOcid"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(AccessConfigFieldsProviderOracleCloud, {
  getInitialValues,
  getSchema,
});

export default _default;
