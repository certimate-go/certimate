import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForHuaweiCloud } from "@/domain/access";

type AccessFormHuaweiCloudConfigFieldValues = Nullish<AccessConfigForHuaweiCloud>;

export type AccessFormHuaweiCloudConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormHuaweiCloudConfigFieldValues;
  onValuesChange?: (values: AccessFormHuaweiCloudConfigFieldValues) => void;
};

const initFormModel = (): AccessFormHuaweiCloudConfigFieldValues => {
  return {
    accessKeyId: "",
    secretAccessKey: "",
  };
};

const AccessFormHuaweiCloudConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormHuaweiCloudConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessKeyId: z
      .string()
      .min(1, t("access.form.huaweicloud_access_key_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    secretAccessKey: z
      .string()
      .min(1, t("access.form.huaweicloud_secret_access_key.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    enterpriseProjectId: z.string().nullish(),
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
        label={t("access.form.huaweicloud_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.huaweicloud_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.huaweicloud_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretAccessKey"
        label={t("access.form.huaweicloud_secret_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.huaweicloud_secret_access_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.huaweicloud_secret_access_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="enterpriseProjectId"
        label={t("access.form.huaweicloud_enterprise_project_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.huaweicloud_enterprise_project_id.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.huaweicloud_enterprise_project_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormHuaweiCloudConfig;
