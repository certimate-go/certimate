import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForAliyun } from "@/domain/access";

type AccessFormAliyunConfigFieldValues = Nullish<AccessConfigForAliyun>;

export type AccessFormAliyunConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormAliyunConfigFieldValues;
  onValuesChange?: (values: AccessFormAliyunConfigFieldValues) => void;
};

const initFormModel = (): AccessFormAliyunConfigFieldValues => {
  return {
    accessKeyId: "",
    accessKeySecret: "",
  };
};

const AccessFormAliyunConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange: onValuesChange }: AccessFormAliyunConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    accessKeyId: z
      .string()
      .min(1, t("access.form.aliyun_access_key_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    accessKeySecret: z
      .string()
      .min(1, t("access.form.aliyun_access_key_secret.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    resourceGroupId: z.string().nullish(),
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
        label={t("access.form.aliyun_access_key_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.aliyun_access_key_id.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.aliyun_access_key_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="accessKeySecret"
        label={t("access.form.aliyun_access_key_secret.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.aliyun_access_key_secret.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.aliyun_access_key_secret.placeholder")} />
      </Form.Item>

      <Form.Item
        name="resourceGroupId"
        label={t("access.form.aliyun_resource_group_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.aliyun_resource_group_id.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.aliyun_resource_group_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormAliyunConfig;
