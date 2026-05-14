import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizApplyNodeConfigFieldsProviderAWSLightsail = () => {
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
        label={t("workflow_node.apply.form.aws_lightsail_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.apply.form.aws_lightsail_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.apply.form.aws_lightsail_region.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "us-east-1",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(t("workflow_node.apply.form.aws_lightsail_region.placeholder")),
  });
};

const _default = Object.assign(BizApplyNodeConfigFieldsProviderAWSLightsail, {
  getInitialValues,
  getSchema,
});

export default _default;
