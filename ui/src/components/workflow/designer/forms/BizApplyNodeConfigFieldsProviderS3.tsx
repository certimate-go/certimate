import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizApplyNodeConfigFieldsProviderS3 = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item name={[parentNamePath, "region"]} initialValue={initialValues.region} label={t("workflow_node.apply.form.s3_region.label")} rules={[formRule]}>
        <Input placeholder={t("workflow_node.apply.form.s3_region.placeholder")} />
      </Form.Item>

      <Form.Item name={[parentNamePath, "bucket"]} initialValue={initialValues.bucket} label={t("workflow_node.apply.form.s3_bucket.label")} rules={[formRule]}>
        <Input placeholder={t("workflow_node.apply.form.s3_bucket.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    bucket: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(t("workflow_node.apply.form.s3_region.placeholder")),
    bucket: z.string().nonempty(t("workflow_node.apply.form.s3_bucket.placeholder")),
  });
};

const _default = Object.assign(BizApplyNodeConfigFieldsProviderS3, {
  getInitialValues,
  getSchema,
});

export default _default;
