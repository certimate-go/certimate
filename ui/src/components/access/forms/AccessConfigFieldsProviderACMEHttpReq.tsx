import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderACMEHttpReq = () => {
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
        label={t("access.form.acmehttpreq_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.acmehttpreq_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "mode"]}
        initialValue={initialValues.mode}
        label={t("access.form.acmehttpreq_mode.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_mode.tooltip") }}></span>}
      >
        <Select
          options={[
            { label: "(default)", value: "" },
            { label: "RAW", value: "RAW" },
          ]}
          placeholder={t("access.form.acmehttpreq_mode.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "username"]}
        initialValue={initialValues.username}
        label={t("access.form.acmehttpreq_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_username.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.acmehttpreq_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "password"]}
        initialValue={initialValues.password}
        label={t("access.form.acmehttpreq_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_password.tooltip") }}></span>}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.acmehttpreq_password.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    endpoint: "https://example.com/api/",
    mode: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    endpoint: z.url(t("common.errmsg.url_invalid")),
    mode: z.string().nullish(),
    username: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    password: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderACMEHttpReq, {
  getInitialValues,
  getSchema,
});

export default _default;
