import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderAWSAPIGateway = () => {
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
        label={t("workflow_node.deploy.form.aws_apigateway_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_apigateway_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_apigateway_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.aws_apigateway_domain.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_apigateway_domain.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateSource"]}
        initialValue={initialValues.certificateSource}
        label={t("workflow_node.deploy.form.aws_apigateway_certificate_source.label")}
        rules={[formRule]}
      >
        <Select
          options={["ACM"].map((s) => ({ label: s, value: s }))}
          placeholder={t("workflow_node.deploy.form.aws_apigateway_certificate_source.placeholder")}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    domain: "",
    certificateSource: "ACM",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(),
    domain: z.string().refine((v) => isDomain(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
    certificateSource: z.string().nonempty(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAWSAPIGateway, {
  getInitialValues,
  getSchema,
});

export default _default;
