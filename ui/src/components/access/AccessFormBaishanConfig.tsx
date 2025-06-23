import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForBaishan } from "@/domain/access";

type AccessFormBaishanConfigFieldValues = Nullish<AccessConfigForBaishan>;

export type AccessFormBaishanConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormBaishanConfigFieldValues;
  onValuesChange?: (values: AccessFormBaishanConfigFieldValues) => void;
};

const initFormModel = (): AccessFormBaishanConfigFieldValues => {
  return {
    apiToken: "",
  };
};

const AccessFormBaishanConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormBaishanConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    apiToken: z
      .string()
      .min(1, t("access.form.baishan_api_token.placeholder"))
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
      <Form.Item name="apiToken" label={t("access.form.baishan_api_token.label")} rules={[formRule]}>
        <Input.Password autoComplete="new-password" placeholder={t("access.form.baishan_api_token.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormBaishanConfig;
