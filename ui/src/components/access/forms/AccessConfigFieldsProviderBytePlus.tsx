import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderBytePlus = () => {
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
        name={[parentNamePath, "accessKey"]}
        initialValue={initialValues.accessKey}
        label={t("access.form.byteplus_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.byteplus_access_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.byteplus_access_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretKey"]}
        initialValue={initialValues.secretKey}
        label={t("access.form.byteplus_secret_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.byteplus_secret_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.byteplus_secret_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "projectName"]}
        initialValue={initialValues.projectName}
        label={t("access.form.byteplus_project_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.byteplus_project_name.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.byteplus_project_name.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    accessKey: "",
    secretKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    accessKey: z.string().nonempty(),
    secretKey: z.string().nonempty(),
    projectName: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderBytePlus, {
  getInitialValues,
  getSchema,
});

export default _default;
