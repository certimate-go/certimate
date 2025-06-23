import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForUCloud } from "@/domain/access";

type AccessFormUCloudConfigFieldValues = Nullish<AccessConfigForUCloud>;

export type AccessFormUCloudConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormUCloudConfigFieldValues;
  onValuesChange?: (values: AccessFormUCloudConfigFieldValues) => void;
};

const initFormModel = (): AccessFormUCloudConfigFieldValues => {
  return {
    privateKey: "",
    publicKey: "",
  };
};

const AccessFormUCloudConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormUCloudConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    privateKey: z
      .string()
      .min(1, t("access.form.ucloud_private_key.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    publicKey: z
      .string()
      .min(1, t("access.form.ucloud_public_key.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    projectId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish(),
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
        name="privateKey"
        label={t("access.form.ucloud_private_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ucloud_private_key.tooltip") }}></span>}
      >
        <Input autoComplete="new-password" placeholder={t("access.form.ucloud_private_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="publicKey"
        label={t("access.form.ucloud_public_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ucloud_public_key.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("access.form.ucloud_public_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name="projectId"
        label={t("access.form.ucloud_project_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.ucloud_project_id.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.ucloud_project_id.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormUCloudConfig;
