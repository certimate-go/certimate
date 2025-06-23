import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForWangsu } from "@/domain/access";

type AccessFormWangsuConfigFieldValues = Nullish<AccessConfigForWangsu>;

export type AccessFormWangsuConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormWangsuConfigFieldValues;
  onValuesChange?: (values: AccessFormWangsuConfigFieldValues) => void;
};

const initFormModel = (): AccessFormWangsuConfigFieldValues => {
  return {
    accessKeyId: "",
    accessKeySecret: "",
    apiKey: "",
  };
};

const AccessFormWangsuConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange: onValuesChange }: AccessFormWangsuConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessKeyId: z
      .string()
      .min(1, t("access.form.wangsu_access_key_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    accessKeySecret: z
      .string()
      .min(1, t("access.form.wangsu_access_key_secret.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    apiKey: z
      .string()
      .min(1, t("access.form.wangsu_api_key.placeholder"))
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
        name="accessKeyId"
        label={t("access.form.wangsu_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.wangsu_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.wangsu_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="accessKeySecret"
        label={t("access.form.wangsu_access_key_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.wangsu_access_key_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.wangsu_access_key_secret.placeholder")} />
      </Form.Item>

      <Form.Item
        name="apiKey"
        label={t("access.form.wangsu_api_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.wangsu_api_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.wangsu_api_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormWangsuConfig;
