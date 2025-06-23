import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForBaiduCloud } from "@/domain/access";

type AccessFormBaiduCloudConfigFieldValues = Nullish<AccessConfigForBaiduCloud>;

export type AccessFormBaiduCloudConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormBaiduCloudConfigFieldValues;
  onValuesChange?: (values: AccessFormBaiduCloudConfigFieldValues) => void;
};

const initFormModel = (): AccessFormBaiduCloudConfigFieldValues => {
  return {
    accessKeyId: "",
    secretAccessKey: "",
  };
};

const AccessFormBaiduCloudConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormBaiduCloudConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessKeyId: z
      .string()
      .min(1, t("access.form.baiducloud_access_key_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    secretAccessKey: z
      .string()
      .min(1, t("access.form.baiducloud_secret_access_key.placeholder"))
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
        label={t("access.form.baiducloud_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.baiducloud_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.baiducloud_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretAccessKey"
        label={t("access.form.baiducloud_secret_access_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.baiducloud_secret_access_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.baiducloud_secret_access_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormBaiduCloudConfig;
