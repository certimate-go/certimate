import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForDeSEC } from "@/domain/access";

type AccessFormDeSECConfigFieldValues = Nullish<AccessConfigForDeSEC>;

export type AccessFormDeSECConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormDeSECConfigFieldValues;
  onValuesChange?: (values: AccessFormDeSECConfigFieldValues) => void;
};

const initFormModel = (): AccessFormDeSECConfigFieldValues => {
  return {
    token: "",
  };
};

const AccessFormDeSECConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormDeSECConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    token: z
      .string()
      .min(1, t("access.form.desec_token.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
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
        name="token"
        label={t("access.form.desec_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.desec_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.desec_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormDeSECConfig;
