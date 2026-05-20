import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderUCloud = () => {
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
        name={[parentNamePath, "privateKey"]}
        initialValue={initialValues.privateKey}
        label={t("access.form.ucloud_private_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ucloud_private_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.ucloud_private_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "publicKey"]}
        initialValue={initialValues.publicKey}
        label={t("access.form.ucloud_public_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ucloud_public_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ucloud_public_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "projectId"]}
        initialValue={initialValues.projectId}
        label={t("access.form.ucloud_project_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ucloud_project_id.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.ucloud_project_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    privateKey: "",
    publicKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    privateKey: z.string().nonempty(),
    publicKey: z.string().nonempty(),
    projectId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderUCloud, {
  getInitialValues,
  getSchema,
});

export default _default;
