import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizApplyNodeConfigFieldsProviderLocal = () => {
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
        name={[parentNamePath, "webRootPath"]}
        initialValue={initialValues.webRootPath}
        label={t("workflow_node.apply.form.local_webroot_path.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.local_webroot_path.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.local_webroot_path.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    webRootPath: "/var/www/html/",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    webRootPath: z
      .string()
      .nonempty()
      .refine((v) => !!v && (v.endsWith("/") || v.endsWith("\\")), t("workflow_node.apply.form.local_webroot_path.placeholder")),
  });
};

const _default = Object.assign(BizApplyNodeConfigFieldsProviderLocal, {
  getInitialValues,
  getSchema,
});

export default _default;
