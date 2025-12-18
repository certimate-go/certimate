import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderKubernetesSecret = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const handleSecretAnnotationsBlur = () => {
    let value = formInst.getFieldValue([parentNamePath, "secretAnnotations"]);
    value = value.trim();
    value = value.replace(/(?<!\r)\n/g, "\r\n");
    formInst.setFieldValue([parentNamePath, "secretAnnotations"], value);
  };

  const handleSecretLabelsBlur = () => {
    let value = formInst.getFieldValue([parentNamePath, "secretLabels"]);
    value = value.trim();
    value = value.replace(/(?<!\r)\n/g, "\r\n");
    formInst.setFieldValue([parentNamePath, "secretLabels"], value);
  };

  return (
    <>
      <Form.Item
        name={[parentNamePath, "namespace"]}
        initialValue={initialValues.namespace}
        label={t("workflow_node.deploy.form.k8s_namespace.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_namespace.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_namespace.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretName"]}
        initialValue={initialValues.secretName}
        label={t("workflow_node.deploy.form.k8s_secret_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretType"]}
        initialValue={initialValues.secretType}
        label={t("workflow_node.deploy.form.k8s_secret_type.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_type.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_type.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretDataKeyForCrt"]}
        initialValue={initialValues.secretDataKeyForCrt}
        label={t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretDataKeyForKey"]}
        initialValue={initialValues.secretDataKeyForKey}
        label={t("workflow_node.deploy.form.k8s_secret_data_key_for_key.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_data_key_for_key.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.k8s_secret_data_key_for_key.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretAnnotations"]}
        initialValue={initialValues.secretAnnotations}
        label={t("workflow_node.deploy.form.k8s_secret_annotations.label")}
        extra={t("workflow_node.deploy.form.k8s_secret_annotations.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_annotations.tooltip") }}></span>}
      >
        <CodeTextInput
          lineWrapping={false}
          height="auto"
          minHeight="64px"
          maxHeight="256px"
          placeholder={t("workflow_node.deploy.form.k8s_secret_annotations.placeholder")}
          onBlur={handleSecretAnnotationsBlur}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "secretLabels"]}
        initialValue={initialValues.secretLabels}
        label={t("workflow_node.deploy.form.k8s_secret_labels.label")}
        extra={t("workflow_node.deploy.form.k8s_secret_labels.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.k8s_secret_labels.tooltip") }}></span>}
      >
        <CodeTextInput
          lineWrapping={false}
          height="auto"
          minHeight="64px"
          maxHeight="256px"
          placeholder={t("workflow_node.deploy.form.k8s_secret_labels.placeholder")}
          onBlur={handleSecretLabelsBlur}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    namespace: "default",
    secretType: "kubernetes.io/tls",
    secretDataKeyForCrt: "tls.crt",
    secretDataKeyForKey: "tls.key",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    namespace: z.string().nonempty(t("workflow_node.deploy.form.k8s_namespace.placeholder")),
    secretName: z.string().nonempty(t("workflow_node.deploy.form.k8s_secret_name.placeholder")),
    secretType: z.string().nonempty(t("workflow_node.deploy.form.k8s_secret_type.placeholder")),
    secretDataKeyForCrt: z.string().nonempty(t("workflow_node.deploy.form.k8s_secret_data_key_for_crt.placeholder")),
    secretDataKeyForKey: z.string().nonempty(t("workflow_node.deploy.form.k8s_secret_data_key_for_key.placeholder")),
    secretAnnotations: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;

        const lines = v.split(/\r?\n/);
        for (const line of lines) {
          if (line.split(":").length < 2) {
            return false;
          }
        }
        return true;
      }, t("workflow_node.deploy.form.k8s_secret_annotations.errmsg.invalid")),
    secretLabels: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;

        const lines = v.split(/\r?\n/);
        for (const line of lines) {
          if (line.split(":").length < 2) {
            return false;
          }
        }
        return true;
      }, t("workflow_node.deploy.form.k8s_secret_labels.errmsg.invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderKubernetesSecret, {
  getInitialValues,
  getSchema,
});

export default _default;
