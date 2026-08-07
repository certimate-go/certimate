import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_METHOD_ACCESSKEY = "accesskey" as const;
const AUTH_METHOD_IMDS = "imds" as const;

const AccessConfigFormFieldsProviderAWS = () => {
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
        label={t("access.form.aws_auth_method.label")}
        rules={[formRule]}
      >
        <Radio.Group block>
          <Radio.Button value={AUTH_METHOD_ACCESSKEY}>{t("access.form.aws_auth_method.option.accesskey.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_IMDS}>{t("access.form.aws_auth_method.option.imds.label")}</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_ACCESSKEY}>
        <Form.Item
          name={[parentNamePath, "accessKeyId"]}
          initialValue={initialValues.accessKeyId}
          label={t("access.form.aws_access_key_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.aws_access_key_id.tooltip") }}></span>}
        >
          <Input autoComplete="new-password" placeholder={t("access.form.aws_access_key_id.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "secretAccessKey"]}
          initialValue={initialValues.secretAccessKey}
          label={t("access.form.aws_secret_access_key.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.aws_secret_access_key.tooltip") }}></span>}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.aws_secret_access_key.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldAuthMethod === AUTH_METHOD_IMDS}>
        <Form.Item>
          <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.aws_auth_method.option.imds.guide") }}></span>} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    authMethod: AUTH_METHOD_ACCESSKEY,
    accessKeyId: "",
    secretAccessKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      authMethod: z.enum([AUTH_METHOD_ACCESSKEY, AUTH_METHOD_IMDS]),
      accessKeyId: z.string().nullish(),
      secretAccessKey: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.authMethod) {
        case AUTH_METHOD_ACCESSKEY:
          {
            const scAccessKeyId = z.string().nonempty();
            const spAccessKeyId = scAccessKeyId.safeParse(values.accessKeyId);
            if (!spAccessKeyId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spAccessKeyId.error).errors.join(),
                path: ["accessKeyId"],
              });
            }

            const scSecretAccessKey = z.string().nonempty();
            const spSecretAccessKey = scSecretAccessKey.safeParse(values.secretAccessKey);
            if (!spSecretAccessKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spSecretAccessKey.error).errors.join(),
                path: ["secretAccessKey"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderAWS, {
  getInitialValues,
  getSchema,
});

export default _default;
