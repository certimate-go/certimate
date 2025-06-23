import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

type DeployNodeConfigFormKubernetesSecretConfigFieldValues = Nullish<{
  namespace: string;
  secretName: string;
  secretType: string;
  secretDataKeyForCrt: string;
  secretDataKeyForKey: string;
}>;

export type DeployNodeConfigFormKubernetesSecretConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormKubernetesSecretConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormKubernetesSecretConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormKubernetesSecretConfigFieldValues => {
  return {
    namespace: "default",
    secretType: "kubernetes.io/tls",
    secretDataKeyForCrt: "tls.crt",
    secretDataKeyForKey: "tls.key",
  };
};

const DeployNodeConfigFormKubernetesSecretConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormKubernetesSecretConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    namespace: z
      .string(t("workflow_node.deploy.form.k8s_namespace.placeholder"))
      .nonempty(t("workflow_node.deploy.form.k8s_namespace.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    secretName: z
      .string(t("workflow_node.deploy.form.k8s_secret_name.placeholder"))
      .nonempty(t("workflow_node.deploy.form.k8s_secret_name.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    secretType: z
      .string(t("workflow_node.deploy.form.k8s_secret_type.placeholder"))
      .nonempty(t("workflow_node.deploy.form.k8s_secret_type.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    secretDataKeyForCrt: z
      .string(t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.placeholder"))
      .nonempty(t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    secretDataKeyForKey: z
      .string(t("workflow_node.deploy.form.k8s_secret_data_key_for_key.placeholder"))
      .nonempty(t("workflow_node.deploy.form.k8s_secret_data_key_for_key.placeholder"))
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
        name="namespace"
        label={t("workflow_node.deploy.form.k8s_namespace.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_namespace.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_namespace.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretName"
        label={t("workflow_node.deploy.form.k8s_secret_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretType"
        label={t("workflow_node.deploy.form.k8s_secret_type.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_type.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_type.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretDataKeyForCrt"
        label={t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.placeholder")} />
      </Form.Item>

      <Form.Item
        name="secretDataKeyForKey"
        label={t("workflow_node.deploy.form.k8s_secret_data_key_for_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_data_key_for_key.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_data_key_for_key.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormKubernetesSecretConfig;
