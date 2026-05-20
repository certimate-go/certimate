import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderGcoreCDN = () => {
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
        name={[parentNamePath, "resourceId"]}
        initialValue={initialValues.resourceId}
        label={t("workflow_node.deploy.form.gcore_cdn_resource_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.gcore_cdn_resource_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("workflow_node.deploy.form.gcore_cdn_resource_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateId"]}
        initialValue={initialValues.certificateId}
        label={t("workflow_node.deploy.form.gcore_cdn_certificate_id.label")}
        extra={t("workflow_node.deploy.form.gcore_cdn_certificate_id.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.gcore_cdn_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear type="number" placeholder={t("workflow_node.deploy.form.gcore_cdn_certificate_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    resourceId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    resourceId: z.union([z.string().nonempty(), z.int().positive()]),
    certificateId: z.union([z.string(), z.int().positive()]).nullish(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderGcoreCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
