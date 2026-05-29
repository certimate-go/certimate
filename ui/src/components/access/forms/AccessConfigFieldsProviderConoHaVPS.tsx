import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const API_VERSION_V2 = "v2" as const;
const API_VERSION_V3 = "v3" as const;

const AccessConfigFormFieldsProviderConoHaVPS = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance<z.infer<typeof formSchema>>();
  const initialValues = getInitialValues();

  const fieldApiVersion = Form.useWatch([parentNamePath, "apiVersion"], formInst);

  return (
    <>
      <Form.Item
        name={[parentNamePath, "apiVersion"]}
        initialValue={initialValues.apiVersion}
        label={t("access.form.conohavps_api_version.label")}
        rules={[formRule]}
      >
        <Select
          options={[API_VERSION_V2, API_VERSION_V3].map((s) => ({ label: s, value: s }))}
          placeholder={t("access.form.conohavps_api_version.placeholder")}
        />
      </Form.Item>

      <Show>
        <Show.Case when={fieldApiVersion === API_VERSION_V2}>
          <Form.Item
            name={[parentNamePath, "apiUserName"]}
            initialValue={initialValues.apiUserName}
            label={t("access.form.conohavps_api_v2_username.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v2_username.tooltip") }}></span>}
          >
            <Input autoComplete="new-password" placeholder={t("access.form.conohavps_api_v2_username.placeholder")} />
          </Form.Item>

          <Form.Item
            name={[parentNamePath, "apiPassword"]}
            initialValue={initialValues.apiPassword}
            label={t("access.form.conohavps_api_v2_password.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v2_password.tooltip") }}></span>}
          >
            <Input.Password autoComplete="new-password" placeholder={t("access.form.conohavps_api_v2_password.placeholder")} />
          </Form.Item>

          <Form.Item
            name={[parentNamePath, "tenantId"]}
            initialValue={initialValues.tenantId}
            label={t("access.form.conohavps_api_v2_tenant_id.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v2_tenant_id.tooltip") }}></span>}
          >
            <Input placeholder={t("access.form.conohavps_api_v2_tenant_id.placeholder")} />
          </Form.Item>
        </Show.Case>

        <Show.Case when={fieldApiVersion === API_VERSION_V3}>
          <Form.Item
            name={[parentNamePath, "apiUserId"]}
            initialValue={initialValues.apiUserId}
            label={t("access.form.conohavps_api_v3_user_id.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v3_user_id.tooltip") }}></span>}
          >
            <Input autoComplete="new-password" placeholder={t("access.form.conohavps_api_v3_user_id.placeholder")} />
          </Form.Item>

          <Form.Item
            name={[parentNamePath, "apiUserName"]}
            initialValue={initialValues.apiUserName}
            label={t("access.form.conohavps_api_v3_user_name.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v3_user_name.tooltip") }}></span>}
          >
            <Input autoComplete="new-password" placeholder={t("access.form.conohavps_api_v3_user_name.placeholder")} />
          </Form.Item>

          <Form.Item
            name={[parentNamePath, "apiPassword"]}
            initialValue={initialValues.apiPassword}
            label={t("access.form.conohavps_api_v3_password.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v3_password.tooltip") }}></span>}
          >
            <Input.Password autoComplete="new-password" placeholder={t("access.form.conohavps_api_v3_password.placeholder")} />
          </Form.Item>

          <Form.Item
            name={[parentNamePath, "tenantId"]}
            initialValue={initialValues.tenantId}
            label={t("access.form.conohavps_api_v3_project_id.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v3_project_id.tooltip") }}></span>}
          >
            <Input placeholder={t("access.form.conohavps_api_v3_project_id.placeholder")} />
          </Form.Item>

          <Form.Item
            name={[parentNamePath, "tenantName"]}
            initialValue={initialValues.tenantName}
            label={t("access.form.conohavps_api_v3_project_name.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.conohavps_api_v3_project_name.tooltip") }}></span>}
          >
            <Input placeholder={t("access.form.conohavps_api_v3_project_name.placeholder")} />
          </Form.Item>
        </Show.Case>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    apiVersion: API_VERSION_V3,
    apiUserId: "",
    apiUserName: "",
    apiPassword: "",
    tenantId: "",
    tenantName: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      apiVersion: z.enum([API_VERSION_V2, API_VERSION_V3]),
      apiUserId: z.string().nullish(),
      apiUserName: z.string().nullish(),
      apiPassword: z.string().nonempty(),
      tenantId: z.string().nullish(),
      tenantName: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.apiVersion) {
        case API_VERSION_V2:
          {
            const scUserName = z.string().nonempty();
            const spUserName = scUserName.safeParse(values.apiUserName);
            if (!spUserName.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spUserName.error).errors.join(),
                path: ["apiUserName"],
              });
            }

            const scTenantId = z.string().nonempty();
            const spTenantId = scTenantId.safeParse(values.tenantId);
            if (!spTenantId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spTenantId.error).errors.join(),
                path: ["tenantId"],
              });
            }
          }
          break;

        case API_VERSION_V3:
          {
            const scUserId = z.string().nonempty();
            const spUserId = scUserId.safeParse(values.apiUserId);
            const scUserName = z.string().nonempty();
            const spUserName = scUserName.safeParse(values.apiUserName);
            if (!spUserId.success && !spUserName.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spUserName.error).errors.join(),
                path: ["apiUserName"],
              });
            } else if (spUserId.success && spUserName.success) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.conohavps_api_v3_user_id.errmsg.conflict"),
                path: ["apiUserId"],
              });
              ctx.addIssue({
                code: "custom",
                message: t("access.form.conohavps_api_v3_user_name.errmsg.conflict"),
                path: ["apiUserName"],
              });
            }

            const scTenantId = z.string().nonempty();
            const spTenantId = scTenantId.safeParse(values.tenantId);
            const scTenantName = z.string().nonempty();
            const spTenantName = scTenantName.safeParse(values.tenantName);
            if (!spTenantId.success && !spTenantName.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spTenantId.error).errors.join(),
                path: ["tenantId"],
              });
            } else if (spTenantId.success && spTenantName.success) {
              ctx.addIssue({
                code: "custom",
                message: t("access.form.conohavps_api_v3_project_id.errmsg.conflict"),
                path: ["tenantId"],
              });
              ctx.addIssue({
                code: "custom",
                message: t("access.form.conohavps_api_v3_project_name.errmsg.conflict"),
                path: ["tenantName"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(AccessConfigFormFieldsProviderConoHaVPS, {
  getInitialValues,
  getSchema,
});

export default _default;
