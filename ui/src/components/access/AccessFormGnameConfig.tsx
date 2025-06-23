import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForGname } from "@/domain/access";

type AccessFormGnameConfigFieldValues = Nullish<AccessConfigForGname>;

export type AccessFormGnameConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormGnameConfigFieldValues;
  onValuesChange?: (values: AccessFormGnameConfigFieldValues) => void;
};

const initFormModel = (): AccessFormGnameConfigFieldValues => {
  return {
    appId: "",
    appKey: "",
  };
};

const AccessFormGnameConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormGnameConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    appId: z
      .string()
      .min(1, t("access.form.gname_app_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    appKey: z
      .string()
      .min(1, t("access.form.gname_app_key.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
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
        name="appId"
        label={t("access.form.gname_app_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.gname_app_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.gname_app_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="appKey"
        label={t("access.form.gname_app_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.gname_app_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.gname_app_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormGnameConfig;
