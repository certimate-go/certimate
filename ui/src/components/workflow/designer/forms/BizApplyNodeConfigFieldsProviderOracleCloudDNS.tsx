import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizApplyNodeConfigFieldsProviderOracleCloudDNS = () => {
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
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.apply.form.oraclecloud_dns_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.oraclecloud_dns_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.oraclecloud_dns_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "compartmentOcid"]}
        initialValue={initialValues.compartmentOcid}
        label={t("workflow_node.apply.form.oraclecloud_dns_compartment_ocid.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.oraclecloud_dns_compartment_ocid.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.oraclecloud_dns_compartment_ocid.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "us-phoenix-1",
    compartmentOcid: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    region: z.string().nonempty(),
    compartmentOcid: z.string().regex(/^ocid\d\..{1,}$/),
  });
};

const _default = Object.assign(BizApplyNodeConfigFieldsProviderOracleCloudDNS, {
  getInitialValues,
  getSchema,
});

export default _default;
