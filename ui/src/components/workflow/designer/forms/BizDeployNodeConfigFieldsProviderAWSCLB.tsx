import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isPortNumber } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderAWSCLB = () => {
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
        label={t("workflow_node.deploy.form.aws_clb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_clb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_clb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "loadbalancerName"]}
        initialValue={initialValues.loadbalancerName}
        label={t("workflow_node.deploy.form.aws_clb_loadbalancer_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_clb_loadbalancer_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aws_clb_loadbalancer_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "loadbalancerPort"]}
        initialValue={initialValues.loadbalancerPort}
        label={t("workflow_node.deploy.form.aws_clb_loadbalancer_port.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aws_clb_loadbalancer_port.tooltip") }}></span>}
      >
        <Input type="number" min={1} max={65535} placeholder={t("workflow_node.deploy.form.aws_clb_loadbalancer_port.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateSource"]}
        initialValue={initialValues.certificateSource}
        label={t("workflow_node.deploy.form.aws_clb_certificate_source.label")}
        rules={[formRule]}
      >
        <Select
          options={["ACM", "IAM"].map((s) => ({ label: s, value: s }))}
          placeholder={t("workflow_node.deploy.form.aws_clb_certificate_source.placeholder")}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    loadbalancerName: "",
    loadbalancerPort: 443,
    certificateSource: "ACM",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    region: z.string().nonempty(),
    loadbalancerName: z.string().nonempty(),
    loadbalancerPort: z.coerce.number().refine((v) => isPortNumber(v), t("common.errmsg.port_invalid")),
    certificateSource: z.string().nonempty(),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAWSCLB, {
  getInitialValues,
  getSchema,
});

export default _default;
