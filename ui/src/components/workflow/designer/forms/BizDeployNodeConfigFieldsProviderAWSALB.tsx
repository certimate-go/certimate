import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderAWSALB = () => {
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
        label={t("workflow_node.deploy.form.aws_alb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_alb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_alb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "loadbalancerArn"]}
        initialValue={initialValues.loadbalancerArn}
        label={t("workflow_node.deploy.form.aws_alb_loadbalancer_arn.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_alb_loadbalancer_arn.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_alb_loadbalancer_arn.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "listenerArn"]}
        initialValue={initialValues.listenerArn}
        label={t("workflow_node.deploy.form.aws_alb_listener_arn.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_alb_listener_arn.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_alb_listener_arn.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateSource"]}
        initialValue={initialValues.certificateSource}
        label={t("workflow_node.deploy.form.aws_alb_certificate_source.label")}
        rules={[formRule]}
      >
        <Select
          options={["ACM", "IAM"].map((s) => ({ label: s, value: s }))}
          placeholder={t("workflow_node.deploy.form.aws_alb_certificate_source.placeholder")}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    loadbalancerArn: "",
    listenerArn: "",
    certificateSource: "ACM",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z.object({
    region: z.string().nonempty(),
    loadbalancerArn: z.string().nonempty(),
    listenerArn: z.string().nonempty(),
    certificateSource: z.string().nonempty(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAWSALB, {
  getInitialValues,
  getSchema,
});

export default _default;
