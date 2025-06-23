import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type AccessConfigForACMEHttpReq } from "@/domain/access";

type AccessFormACMEHttpReqConfigFieldValues = Nullish<AccessConfigForACMEHttpReq>;

export type AccessFormACMEHttpReqConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormACMEHttpReqConfigFieldValues;
  onValuesChange?: (values: AccessFormACMEHttpReqConfigFieldValues) => void;
};

const initFormModel = (): AccessFormACMEHttpReqConfigFieldValues => {
  return {
    endpoint: "https://example.com/api/",
    mode: "",
  };
};

const AccessFormACMEHttpReqConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormACMEHttpReqConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.url(t("common.errmsg.url_invalid")),
    mode: z.string().nullish(),
    username: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    password: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
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
        name="endpoint"
        label={t("access.form.acmehttpreq_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_endpoint.tooltip") }}></span>}
      >
        <Input placeholder={t("access.form.acmehttpreq_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="mode"
        label={t("access.form.acmehttpreq_mode.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_mode.tooltip") }}></span>}
      >
        <Select
          options={[
            { value: "", label: "(default)" },
            { value: "RAW", label: "RAW" },
          ]}
          placeholder={t("access.form.acmehttpreq_mode.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name="username"
        label={t("access.form.acmehttpreq_username.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_username.tooltip") }}></span>}
      >
        <Input allowClear autoComplete="new-password" placeholder={t("access.form.acmehttpreq_username.placeholder")} />
      </Form.Item>

      <Form.Item
        name="password"
        label={t("access.form.acmehttpreq_password.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.acmehttpreq_password.tooltip") }}></span>}
      >
        <Input.Password allowClear autoComplete="new-password" placeholder={t("access.form.acmehttpreq_password.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormACMEHttpReqConfig;
