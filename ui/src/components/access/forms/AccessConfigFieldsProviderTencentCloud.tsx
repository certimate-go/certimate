import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderTencentCloud = () => {
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
        name={[parentNamePath, "secretId"]}
        initialValue={initialValues.secretId}
        label={t("access.form.tencentcloud_secret_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.tencentcloud_secret_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.tencentcloud_secret_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretKey"]}
        initialValue={initialValues.secretKey}
        label={t("access.form.tencentcloud_secret_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.tencentcloud_secret_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.tencentcloud_secret_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "projectId"]}
        initialValue={initialValues.projectId}
        label={t("access.form.tencentcloud_project_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.tencentcloud_project_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.tencentcloud_project_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    secretId: "",
    secretKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    secretId: z.string().nonempty(),
    secretKey: z.string().nonempty(),
    projectId: z.coerce.number().int().positive().or(z.literal("")).nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderTencentCloud, {
  getInitialValues,
  getSchema,
});

export default _default;
