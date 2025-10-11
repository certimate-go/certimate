import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import { validDomainName } from "@/utils/validators";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";
const MATCH_PATTERN_EXACT = "exact" as const;
const MATCH_PATTERN_WILDCARD = "wildcard" as const;

const BizDeployNodeConfigFieldsProviderTencentCloudEO = () => {
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
        name={[parentNamePath, "endpoint"]}
        initialValue={initialValues.endpoint}
        label={t("workflow_node.deploy.form.tencentcloud_eo_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_eo_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "zoneId"]}
        initialValue={initialValues.zoneId}
        label={t("workflow_node.deploy.form.tencentcloud_eo_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_eo_zone_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "matchPattern"]}
        initialValue={initialValues.matchPattern}
        label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
        extra={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_domain_match_pattern.help_wildcard") }}></span>}
        rules={[formRule]}
      >
        <Radio.Group
          options={[MATCH_PATTERN_EXACT, MATCH_PATTERN_WILDCARD].map((s) => ({
            key: s,
            label: t(`workflow_node.deploy.form.shared_domain_match_pattern.option.${s}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domains"]}
        initialValue={initialValues.domains}
        label={t("workflow_node.deploy.form.tencentcloud_eo_domains.label")}
        extra={t("workflow_node.deploy.form.tencentcloud_eo_domains.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eo_domains.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.tencentcloud_eo_domains.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.tencentcloud_eo_domains.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.tencentcloud_eo_domains.multiple_input_modal.placeholder")}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    matchPattern: MATCH_PATTERN_EXACT,
    zoneId: "",
    domains: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    endpoint: z.string().nullish(),
    zoneId: z.string().nonempty(t("workflow_node.deploy.form.tencentcloud_eo_zone_id.placeholder")),
    matchPattern: z.enum(
      [MATCH_PATTERN_EXACT, MATCH_PATTERN_WILDCARD],
      t("workflow_node.deploy.form.shared_domain_match_pattern.placeholder")
    ),
    domains: z.string().refine((v) => {
      if (!v) return false;
      return String(v)
        .split(MULTIPLE_INPUT_SEPARATOR)
        .every((e) => validDomainName(e, { allowWildcard: true }));
    }, t("common.errmsg.domain_invalid")),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderTencentCloudEO, {
  getInitialValues,
  getSchema,
});

export default _default;
