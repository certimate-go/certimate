import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderHuaweiCloud = () => {
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
        label={t("access.form.huaweicloud_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.huaweicloud_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.huaweicloud_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretAccessKey"]}
        initialValue={initialValues.secretAccessKey}
        label={t("access.form.huaweicloud_secret_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.huaweicloud_secret_access_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.huaweicloud_secret_access_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "enterpriseProjectId"]}
        initialValue={initialValues.enterpriseProjectId}
        label={t("access.form.huaweicloud_enterprise_project_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.huaweicloud_enterprise_project_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.huaweicloud_enterprise_project_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    accessKeyId: "",
    secretAccessKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    accessKeyId: z.string().nonempty(),
    secretAccessKey: z.string().nonempty(),
    enterpriseProjectId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderHuaweiCloud, {
  getInitialValues,
  getSchema,
});

export default _default;
