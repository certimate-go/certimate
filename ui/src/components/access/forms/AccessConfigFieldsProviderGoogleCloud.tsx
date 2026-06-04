import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import FileTextInput from "@/components/FileTextInput";
import { isJsonObject } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFieldsProviderGoogleCloud = () => {
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
        name={[parentNamePath, "projectId"]}
        initialValue={initialValues.projectId}
        label={t("access.form.googlecloud_project_id.label")}
        rules={[formRule]}
      >
        <Input type="url" placeholder={t("access.form.googlecloud_project_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "serviceAccountKey"]}
        initialValue={initialValues.serviceAccountKey}
        label={t("access.form.googlecloud_service_account_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.googlecloud_service_account_key.tooltip") }}></span>}
      >
        <FileTextInput autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("access.form.googlecloud_service_account_key.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    projectId: "",
    serviceAccountKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    projectId: z.string().nonempty(),
    serviceAccountKey: z.string().refine((v) => isJsonObject(v), t("common.errmsg.json_invalid")),
  });
};

const _default = Object.assign(AccessConfigFieldsProviderGoogleCloud, {
  getInitialValues,
  getSchema,
});

export default _default;
