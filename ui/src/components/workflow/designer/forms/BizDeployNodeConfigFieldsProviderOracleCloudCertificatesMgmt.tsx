import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderOracleCloudCertificatesMgmt = () => {
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
        name={[parentNamePath, "compartmentOcid"]}
        initialValue={initialValues.compartmentOcid}
        label={t("workflow_node.deploy.form.oraclecloud_certificatesmgmt_compartment_ocid.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.oraclecloud_certificatesmgmt_compartment_ocid.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.oraclecloud_certificatesmgmt_compartment_ocid.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    compartmentOcid: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    compartmentOcid: z.string().regex(/^ocid\d\..{1,}$/),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderOracleCloudCertificatesMgmt, {
  getInitialValues,
  getSchema,
});

export default _default;
