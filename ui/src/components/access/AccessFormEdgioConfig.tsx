import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForEdgio } from "@/domain/access";

type AccessFormEdgioConfigFieldValues = Nullish<AccessConfigForEdgio>;

export type AccessFormEdgioConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormEdgioConfigFieldValues;
  onValuesChange?: (values: AccessFormEdgioConfigFieldValues) => void;
};

const initFormModel = (): AccessFormEdgioConfigFieldValues => {
  return {
    clientId: "",
    clientSecret: "",
  };
};

const AccessFormEdgioConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormEdgioConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    clientId: z
      .string()
      .min(1, t("access.form.edgio_client_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    clientSecret: z
      .string()
      .min(1, t("access.form.edgio_client_secret.placeholder"))
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
        name="clientId"
        label={t("access.form.edgio_client_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.edgio_client_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.edgio_client_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="clientSecret"
        label={t("access.form.edgio_client_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.edgio_client_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.edgio_client_secret.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormEdgioConfig;
