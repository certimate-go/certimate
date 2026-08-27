import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizPurgeNodeConfigFieldsProviderAzureKeyVault = () => {
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
        label={t("workflow_node.purge.form.azure_keyvault_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.purge.form.azure_keyvault_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.purge.form.azure_keyvault_name.placeholder")} />
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
  const { t: _ } = i18n;

  return z.object({
    keyvaultName: z.string().nonempty(),
  });
};

const _default = Object.assign(BizPurgeNodeConfigFieldsProviderAzureKeyVault, {
  getInitialValues,
  getSchema,
});

export default _default;
