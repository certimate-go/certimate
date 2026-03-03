import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderTodayNIC = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item name={[parentNamePath, "userId"]} initialValue={initialValues.userId} label={t("access.form.todaynic_user_id.label")} rules={[formRule]}>
        <Input autoComplete="new-password" placeholder={t("access.form.todaynic_user_id.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "apiKey"]} initialValue={initialValues.apiKey} label={t("access.form.todaynic_api_key.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.todaynic_api_key.placeholder")} />
      </Form.Item>

      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("access.form.todaynic_agent.guide") }}></span>} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    userId: "",
    apiKey: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    userId: z.string().nonempty(t("access.form.todaynic_user_id.placeholder")),
    apiKey: z.string().nonempty(t("access.form.todaynic_api_key.placeholder")),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderTodayNIC, {
  getInitialValues,
  getSchema,
});

export default _default;
