import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderAzureKeyVault = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item
        name={[parentNamePath, "keyvaultName"]}
        initialValue={initialValues.keyvaultName}
        label={t("workflow_node.deploy.form.azure_keyvault_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.azure_keyvault_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.azure_keyvault_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateName"]}
        initialValue={initialValues.certificateName}
        label={t("workflow_node.deploy.form.azure_keyvault_certificate_name.label")}
        extra={t("workflow_node.deploy.form.azure_keyvault_certificate_name.help")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.azure_keyvault_certificate_name.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    keyvaultName: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    keyvaultName: z.string().nonempty(),
    certificateName: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return true;
        return /^[a-zA-Z0-9-]{1,127}$/.test(v);
      }, t("workflow_node.deploy.form.azure_keyvault_certificate_name.errmsg.invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAzureKeyVault, {
  getInitialValues,
  getSchema,
});

export default _default;
