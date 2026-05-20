import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderAWSCloudFront = () => {
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
        label={t("workflow_node.deploy.form.aws_cloudfront_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_cloudfront_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_cloudfront_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "distributionId"]}
        initialValue={initialValues.distributionId}
        label={t("workflow_node.deploy.form.aws_cloudfront_distribution_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_cloudfront_distribution_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_cloudfront_distribution_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateSource"]}
        initialValue={initialValues.certificateSource}
        label={t("workflow_node.deploy.form.aws_cloudfront_certificate_source.label")}
        rules={[formRule]}
      >
        <Select placeholder={t("workflow_node.deploy.form.aws_cloudfront_certificate_source.placeholder")}>
          <Select.Option key="ACM" value="ACM">
            ACM
          </Select.Option>
          <Select.Option key="IAM" value="IAM">
            IAM
          </Select.Option>
        </Select>
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    certificateSource: "ACM",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    region: z.string().nonempty(),
    distributionId: z.string().nonempty(),
    certificateSource: z.string().nonempty(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAWSCloudFront, {
  getInitialValues,
  getSchema,
});

export default _default;
