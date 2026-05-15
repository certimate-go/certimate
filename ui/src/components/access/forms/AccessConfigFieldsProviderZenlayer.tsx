import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderZenlayer = () => {
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
        name={[parentNamePath, "accessKeyId"]}
        initialValue={initialValues.accessKeyId}
        label={t("access.form.zenlayer_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.zenlayer_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.zenlayer_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "accessKeyPassword"]}
        initialValue={initialValues.accessKeyPassword}
        label={t("access.form.zenlayer_access_key_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.zenlayer_access_key_password.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.zenlayer_access_key_password.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceGroupId"]}
        initialValue={initialValues.resourceGroupId}
        label={t("access.form.zenlayer_resource_group_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.zenlayer_resource_group_id.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.zenlayer_resource_group_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    accessKeyId: "",
    accessKeyPassword: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    accessKeyId: z.string().nonempty(t("access.form.zenlayer_access_key_id.placeholder")),
    accessKeyPassword: z.string().nonempty(t("access.form.zenlayer_access_key_password.placeholder")),
    resourceGroupId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderZenlayer, {
  getInitialValues,
  getSchema,
});

export default _default;
