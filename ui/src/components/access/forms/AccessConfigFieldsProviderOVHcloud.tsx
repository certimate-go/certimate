import { getI18n, useTranslation } from "react-i18next";
import { AutoComplete, Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { matchSearchOption } from "@/utils/search";

import { useFormNestedFieldsContext } from "./_context";

const AUTH_METHOD_APPLICATION = "application" as const;
const AUTH_METHOD_OAUTH2 = "oauth2" as const;

const AccessConfigFormFieldsProviderOVHcloud = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldAuthMethod = Form.useWatch<string>([parentNamePath, "authMethod"], formInst);

  return (
    <>
      <Form.Item name={[parentNamePath, "endpoint"]} initialValue={initialValues.endpoint} label={t("access.form.ovhcloud_endpoint.label")} rules={[formRule]}>
        <AutoComplete
          options={["ovh-eu", "ovh-us", "ovh-ca"].map((value) => ({ value }))}
          placeholder={t("access.form.ovhcloud_endpoint.placeholder")}
          showSearch={{
            filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
          }}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "authMethod"]}
        initialValue={initialValues.authMethod}
        label={t("access.form.ovhcloud_auth_method.label")}
        rules={[formRule]}
      >
        <Radio.Group block>
          <Radio.Button value={AUTH_METHOD_APPLICATION}>{t("access.form.ovhcloud_auth_method.option.application.label")}</Radio.Button>
          <Radio.Button value={AUTH_METHOD_OAUTH2}>{t("access.form.ovhcloud_auth_method.option.oauth2.label")}</Radio.Button>
        </Radio.Group>
      </Form.Item>

      <Show when={fieldAuthMethod === AUTH_METHOD_APPLICATION}>
        <Form.Item
          name={[parentNamePath, "applicationKey"]}
          initialValue={initialValues.applicationKey}
          label={t("access.form.ovhcloud_application_key.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ovhcloud_application_key.tooltip") }}></span>}
        >
          <Input autoComplete="new-password" placeholder={t("access.form.ovhcloud_application_key.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "applicationSecret"]}
          initialValue={initialValues.applicationSecret}
          label={t("access.form.ovhcloud_application_secret.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ovhcloud_application_secret.tooltip") }}></span>}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.ovhcloud_application_secret.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "consumerKey"]}
          initialValue={initialValues.consumerKey}
          label={t("access.form.ovhcloud_consumer_key.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ovhcloud_consumer_key.tooltip") }}></span>}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.ovhcloud_consumer_key.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldAuthMethod === AUTH_METHOD_OAUTH2}>
        <Form.Item
          name={[parentNamePath, "clientId"]}
          initialValue={initialValues.clientId}
          label={t("access.form.ovhcloud_client_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ovhcloud_client_id.tooltip") }}></span>}
        >
          <Input autoComplete="new-password" placeholder={t("access.form.ovhcloud_client_id.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "clientSecret"]}
          initialValue={initialValues.clientSecret}
          label={t("access.form.ovhcloud_client_secret.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ovhcloud_client_secret.tooltip") }}></span>}
        >
          <Input.Password autoComplete="new-password" placeholder={t("access.form.ovhcloud_client_secret.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    endpoint: "ovh-eu",
    authMethod: AUTH_METHOD_APPLICATION,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      endpoint: z.string().nonempty(),
      authMethod: z.enum([AUTH_METHOD_APPLICATION, AUTH_METHOD_OAUTH2]),
      applicationKey: z.string().nullish(),
      applicationSecret: z.string().nullish(),
      consumerKey: z.string().nullish(),
      clientId: z.string().nullish(),
      clientSecret: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.authMethod) {
        case AUTH_METHOD_APPLICATION:
          {
            const scApplicationKey = z.string().nonempty();
            const spApplicationKey = scApplicationKey.safeParse(values.applicationKey);
            if (!spApplicationKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spApplicationKey.error).errors.join(),
                path: ["applicationKey"],
              });
            }

            const scApplicationSecret = z.string().nonempty();
            const spApplicationSecret = scApplicationSecret.safeParse(values.applicationSecret);
            if (!spApplicationSecret.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spApplicationSecret.error).errors.join(),
                path: ["applicationSecret"],
              });
            }

            const scConsumerKey = z.string().nonempty();
            const spConsumerKey = scConsumerKey.safeParse(values.consumerKey);
            if (!spConsumerKey.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spConsumerKey.error).errors.join(),
                path: ["consumerKey"],
              });
            }
          }
          break;

        case AUTH_METHOD_OAUTH2:
          {
            const scClientId = z.string().nonempty();
            const spClientId = scClientId.safeParse(values.clientId);
            if (!spClientId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spClientId.error).errors.join(),
                path: ["clientId"],
              });
            }

            const scClientSecret = z.string().nonempty();
            const spClientSecret = scClientSecret.safeParse(values.clientSecret);
            if (!spClientSecret.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spClientSecret.error).errors.join(),
                path: ["clientSecret"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderOVHcloud, {
  getInitialValues,
  getSchema,
});

export default _default;
