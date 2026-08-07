import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;
const DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard" as const;
const DOMAIN_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProviderTencentCloudEOMakers = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldDomainMatchPattern = Form.useWatch([parentNamePath, "domainMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "endpoint"]}
        initialValue={initialValues.endpoint}
        label={t("workflow_node.deploy.form.tencentcloud_eomakers_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eomakers_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_eomakers_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "apiToken"]}
        initialValue={initialValues.apiToken}
        label={t("workflow_node.deploy.form.tencentcloud_eomakers_api_token.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eomakers_api_token.tooltip") }}></span>}
      >
        <Input.Password autoComplete="new-password" placeholder={t("workflow_node.deploy.form.tencentcloud_eomakers_api_token.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "projectId"]}
        initialValue={initialValues.projectId}
        label={t("workflow_node.deploy.form.tencentcloud_eomakers_project_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_eomakers_project_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_eomakers_project_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domainMatchPattern"]}
        initialValue={initialValues.domainMatchPattern}
        label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
        extra={
          fieldDomainMatchPattern === DOMAIN_MATCH_PATTERN_EXACT ? (
            <span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_domain_match_pattern.option.exact.help.wildcard") }}></span>
          ) : (
            void 0
          )
        }
        rules={[formRule]}
      >
        <Radio.Group
          options={[DOMAIN_MATCH_PATTERN_EXACT, DOMAIN_MATCH_PATTERN_WILDCARD, DOMAIN_MATCH_PATTERN_CERTSAN].map((s) => ({
            label: t(`workflow_node.deploy.form.shared_domain_match_pattern.option.${s}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Show when={fieldDomainMatchPattern !== DOMAIN_MATCH_PATTERN_CERTSAN}>
        <Form.Item
          name={[parentNamePath, "domains"]}
          initialValue={initialValues.domains}
          label={t("workflow_node.deploy.form.tencentcloud_eomakers_domains.label")}
          extra={t("workflow_node.deploy.form.tencentcloud_eomakers_domains.help")}
          rules={[formRule]}
        >
          <MultipleSplitValueInput
            modalTitle={t("workflow_node.deploy.form.tencentcloud_eomakers_domains.multiple_input_modal.title")}
            placeholder={t("workflow_node.deploy.form.tencentcloud_eomakers_domains.placeholder")}
            placeholderInModal={t("workflow_node.deploy.form.tencentcloud_eomakers_domains.multiple_input_modal.placeholder")}
            splitOptions={{ removeEmpty: true, trimSpace: true }}
          />
        </Form.Item>
      </Show>

      <Form.Item
        label={t("workflow_node.deploy.form.tencentcloud_eomakers_enable_multiple_ssl.label")}
        extra={t("workflow_node.deploy.form.tencentcloud_eomakers_enable_multiple_ssl.help")}
      >
        <span className="inline-block">
          <Form.Item name={[parentNamePath, "enableMultipleSSL"]} initialValue={initialValues.enableMultipleSSL} noStyle rules={[formRule]}>
            <Switch
              checkedChildren={t("workflow_node.deploy.form.tencentcloud_eomakers_enable_multiple_ssl.switch.on")}
              unCheckedChildren={t("workflow_node.deploy.form.tencentcloud_eomakers_enable_multiple_ssl.switch.off")}
            />
          </Form.Item>
        </span>
        <span className="ms-2 inline-block">{t("workflow_node.deploy.form.tencentcloud_eomakers_enable_multiple_ssl.switch.suffix")}</span>
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    apiToken: "",
    projectId: "",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domains: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      endpoint: z.string().nullish(),
      apiToken: z.string().nonempty(),
      projectId: z.string().nonempty(),
      domainMatchPattern: z.string().nonempty().default(DOMAIN_MATCH_PATTERN_EXACT),
      domains: z.string().nullish(),
      enableMultipleSSL: z.boolean().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern) {
        switch (values.domainMatchPattern) {
          case DOMAIN_MATCH_PATTERN_EXACT:
          case DOMAIN_MATCH_PATTERN_WILDCARD:
            {
              const scDomains = z
                .string()
                .nonempty()
                .refine((v) => v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isDomain(e, { allowWildcard: true })), t("common.errmsg.domain_invalid"));
              const spDomains = scDomains.safeParse(values.domains);
              if (!spDomains.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spDomains.error).errors.join(),
                  path: ["domains"],
                });
              }
            }
            break;
        }
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderTencentCloudEOMakers, {
  getInitialValues,
  getSchema,
});

export default _default;
