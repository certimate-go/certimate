import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForRainYun } from "@/domain/access";

type AccessFormRainYunConfigFieldValues = Nullish<AccessConfigForRainYun>;

export type AccessFormRainYunConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormRainYunConfigFieldValues;
  onValuesChange?: (values: AccessFormRainYunConfigFieldValues) => void;
};

const initFormModel = (): AccessFormRainYunConfigFieldValues => {
  return {
    apiKey: "",
  };
};

const AccessFormRainYunConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormRainYunConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiKey: z
      .string()
      .min(1, t("access.form.rainyun_api_key.placeholder"))
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
        name="apiKey"
        label={t("access.form.rainyun_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.rainyun_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.rainyun_api_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormRainYunConfig;
