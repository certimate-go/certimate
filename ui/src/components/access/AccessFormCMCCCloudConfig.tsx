import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";
import { type AccessConfigForCMCCCloud } from "@/domain/access";

type AccessFormCMCCCloudConfigFieldValues = Nullish<AccessConfigForCMCCCloud>;

export type AccessFormCMCCCloudConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormCMCCCloudConfigFieldValues;
  onValuesChange?: (values: AccessFormCMCCCloudConfigFieldValues) => void;
};

const initFormModel = (): AccessFormCMCCCloudConfigFieldValues => {
  return {
    accessKeyId: "",
    accessKeySecret: "",
  };
};

const AccessFormCMCCCloudConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange: onValuesChange }: AccessFormCMCCCloudConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessKeyId: z
      .string()
      .min(1, t("access.form.cmcccloud_access_key_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    accessKeySecret: z
      .string()
      .min(1, t("access.form.cmcccloud_access_key_secret.placeholder"))
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
        name="accessKeyId"
        label={t("access.form.cmcccloud_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cmcccloud_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.cmcccloud_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="accessKeySecret"
        label={t("access.form.cmcccloud_access_key_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.cmcccloud_access_key_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.cmcccloud_access_key_secret.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormCMCCCloudConfig;
