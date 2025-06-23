import { useTranslation } from "react-i18next";
import { Form, type FormInstance } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import TextFileInput from "@/components/TextFileInput";
import { type AccessConfigForKubernetes } from "@/domain/access";

type AccessFormKubernetesConfigFieldValues = Nullish<AccessConfigForKubernetes>;

export type AccessFormKubernetesConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: AccessFormKubernetesConfigFieldValues;
  onValuesChange?: (values: AccessFormKubernetesConfigFieldValues) => void;
};

const initFormModel = (): AccessFormKubernetesConfigFieldValues => {
  return {};
};

const AccessFormKubernetesConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: AccessFormKubernetesConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    kubeConfig: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
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
        name="kubeConfig"
        label={t("access.form.k8s_kubeconfig.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("access.form.k8s_kubeconfig.tooltip") }}></span>}
      >
        <TextFileInput allowClear autoSize={{ minRows: 3, maxRows: 10 }} placeholder={t("access.form.k8s_kubeconfig.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default AccessFormKubernetesConfig;
