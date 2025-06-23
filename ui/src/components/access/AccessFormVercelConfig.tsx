import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForVercel } from "@/domain/access";

type AccessFormVercelConfigFieldValues = Nullish<AccessConfigForVercel>;

export type AccessFormVercelConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormVercelConfigFieldValues;
  onValuesChange?: (values: AccessFormVercelConfigFieldValues) => void;
};

const initFormModel = (): AccessFormVercelConfigFieldValues => {
  return {
    apiAccessToken: "",
  };
};

const AccessFormVercelConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormVercelConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiAccessToken: z
      .string()
      .min(1, t("access.form.vercel_api_access_token.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    teamId: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
    onValuesChange?.(values);
  };

  return (
    <Form
      form={formInst}
      disabled={disabled}
      initialValues={initialValues ?? initFormModel()}
      layout="vertical"
      name={formName}
      onValuesChange={handleFormChange}
    >
      <Form.Item
        name="apiAccessToken"
        label={t("access.form.vercel_api_access_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.vercel_api_access_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.vercel_api_access_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name="teamId"
        label={t("access.form.vercel_team_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.vercel_team_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("access.form.vercel_team_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormVercelConfig;
