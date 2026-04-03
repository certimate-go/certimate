import { getI18n, useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const BizDeployNodeConfigFieldsProviderJDCloudWAF = () => {
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
        name={[parentNamePath, "regionId"]}
        initialValue={initialValues.regionId}
        label={t("workflow_node.deploy.form.jdcloud_waf_region_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_waf_region_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.jdcloud_waf_region_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "instanceId"]}
        initialValue={initialValues.instanceId}
        label={t("workflow_node.deploy.form.jdcloud_waf_instance_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_waf_instance_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.jdcloud_waf_instance_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.jdcloud_waf_domain.label")}
        rules={[formRule]}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.jdcloud_waf_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    regionId: "",
    instanceId: "",
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    regionId: z.string().nonempty(t("workflow_node.deploy.form.jdcloud_waf_region_id.placeholder")),
    instanceId: z.string().nonempty(t("workflow_node.deploy.form.jdcloud_waf_instance_id.placeholder")),
    domain: z.string().refine((v) => isDomain(v), t("common.errmsg.domain_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderJDCloudWAF, {
  getInitialValues,
  getSchema,
});

export default _default;
