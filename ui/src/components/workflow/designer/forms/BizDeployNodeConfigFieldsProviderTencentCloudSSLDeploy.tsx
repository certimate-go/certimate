import { getI18n, useTranslation } from "react-i18next";
import { AutoComplete, Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Tips from "@/components/Tips";
import { matchSearchOption } from "@/utils/search";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssldeploy.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "endpoint"]}
        initialValue={initialValues.endpoint}
        label={t("workflow_node.deploy.form.tencentcloud_ssldeploy_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssldeploy_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_ssldeploy_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.tencentcloud_ssldeploy_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssldeploy_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_ssldeploy_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceProduct"]}
        initialValue={initialValues.resourceProduct}
        label={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_product.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_product.tooltip") }}></span>}
      >
        <AutoComplete
          options={["apigateway", "cdn", "clb", "cos", "ddos", "lighthouse", "live", "tcb", "teo", "tke", "tse", "vod", "waf"].map((value) => ({ value }))}
          placeholder={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_product.placeholder")}
          showSearch={{
            filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
          }}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceIds"]}
        initialValue={initialValues.resourceIds}
        label={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.label")}
        extra={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.multiple_input_modal.placeholder")}
          separator={MULTIPLE_INPUT_SEPARATOR}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    resourceProduct: "",
    resourceIds: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    endpoint: z.string().nullish(),
    region: z.string().nonempty(),
    resourceProduct: z.string().nonempty(),
    resourceIds: z
      .string()
      .nonempty()
      .refine((v) => {
        return v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => /^[A-Za-z0-9*._\-|]+$/.test(e));
      }, t("workflow_node.deploy.form.tencentcloud_ssldeploy_resource_ids.errmsg.invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy, {
  getInitialValues,
  getSchema,
});

export default _default;
