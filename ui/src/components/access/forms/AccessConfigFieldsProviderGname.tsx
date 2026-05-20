import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderGname = () => {
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
        name={[parentNamePath, "appId"]}
        initialValue={initialValues.appId}
        label={t("access.form.gname_app_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.gname_app_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.gname_app_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "appKey"]}
        initialValue={initialValues.appKey}
        label={t("access.form.gname_app_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.gname_app_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.gname_app_key.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    appId: "",
    appKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    appId: z.string().nonempty(),
    appKey: z.string().nonempty(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderGname, {
  getInitialValues,
  getSchema,
});

export default _default;
