import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isHostname, isUrlWithHttpOrHttps } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFieldsProviderS3 = () => {
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
        label={t("access.form.s3_endpoint.label")}
        extra={<span dangerouslySetInnerHTML={{ __html: t("access.form.s3_endpoint.help") }}></span>}
        rules={[formRule]}
      >
        <Input placeholder={t("access.form.s3_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "accessKey"]} initialValue={initialValues.accessKey} label={t("access.form.s3_access_key.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.s3_access_key.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "secretKey"]} initialValue={initialValues.secretKey} label={t("access.form.s3_secret_key.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.s3_secret_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "signatureVersion"]}
        initialValue={initialValues.signatureVersion}
        label={t("access.form.s3_signature_version.label")}
        rules={[formRule]}
      >
        <Select options={["v2", "v4"].map((s) => ({ label: s, value: s }))} placeholder={t("access.form.s3_signature_version.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "usePathStyle"]}
        initialValue={initialValues.usePathStyle}
        label={t("access.form.s3_use_path_style.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.s3_use_path_style.tooltip") }}></span>}
      >
        <Switch />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "allowInsecureConnections"]}
        initialValue={initialValues.allowInsecureConnections}
        label={t("access.form.shared_allow_insecure_conns.label")}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    endpoint: "",
    accessKey: "",
    secretKey: "",
    signatureVersion: "v4",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    endpoint: z.string().refine((v) => isHostname(v) || isUrlWithHttpOrHttps(v), t("access.form.s3_endpoint.placeholder")),
    accessKey: z.string().nonempty(t("access.form.s3_access_key.placeholder")),
    secretKey: z.string().nonempty(t("access.form.s3_secret_key.placeholder")),
    signatureVersion: z.enum(["v2", "v4"]),
    usePathStyle: z.boolean().nullish(),
    allowInsecureConnections: z.boolean().nullish(),
  });
};

const _default = Object.assign(AccessConfigFieldsProviderS3, {
  getInitialValues,
  getSchema,
});

export default _default;
