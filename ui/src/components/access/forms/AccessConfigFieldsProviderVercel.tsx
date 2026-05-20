import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const AccessConfigFormFieldsProviderVercel = () => {
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
        name={[parentNamePath, "apiAccessToken"]}
        initialValue={initialValues.apiAccessToken}
        label={t("access.form.vercel_api_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.vercel_api_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.vercel_api_access_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "teamId"]}
        initialValue={initialValues.teamId}
        label={t("access.form.vercel_team_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.vercel_team_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.vercel_team_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    apiAccessToken: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    apiAccessToken: z.string().nonempty(),
    teamId: z.string().nullish(),
  });
};

const _default = Object.assign(AccessConfigFormFieldsProviderVercel, {
  getInitialValues,
  getSchema,
});

export default _default;
