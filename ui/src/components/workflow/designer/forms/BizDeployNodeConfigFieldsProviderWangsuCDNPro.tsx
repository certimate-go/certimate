import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;

const BizDeployNodeConfigFieldsProviderWangsuCDNPro = () => {
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
        name={[parentNamePath, "environment"]}
        initialValue={initialValues.environment}
        label={t("workflow_node.deploy.form.wangsu_cdnpro_environment.label")}
        rules={[formRule]}
      >
        <Select
          options={["production", "stating"].map((s) => ({
            label: t(`workflow_node.deploy.form.wangsu_cdnpro_environment.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_environment.placeholder")}
        />
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
          options={[DOMAIN_MATCH_PATTERN_EXACT].map((s) => ({
            label: t(`workflow_node.deploy.form.shared_domain_match_pattern.option.${s}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.wangsu_cdnpro_domain.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_domain.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "certificateId"]}
        initialValue={initialValues.certificateId}
        label={t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.label")}
        extra={t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_certificate_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "webhookId"]}
        initialValue={initialValues.webhookId}
        label={t("workflow_node.deploy.form.wangsu_cdnpro_webhook_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_cdnpro_webhook_id.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.wangsu_cdnpro_webhook_id.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    environment: "production",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      environment: z.enum(["production", "staging"]).default("production"),
      domainMatchPattern: z.string().nonempty().default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().nullish(),
      certificateId: z.string().nullish(),
      webhookId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern) {
        switch (values.domainMatchPattern) {
          case DOMAIN_MATCH_PATTERN_EXACT:
            {
              const scDomain = z.string().refine((v) => isDomain(v, { allowWildcard: true }), t("common.errmsg.domain_invalid"));
              const spDomain = scDomain.safeParse(values.domain);
              if (!spDomain.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spDomain.error).errors.join(),
                  path: ["domain"],
                });
              }
            }
            break;
        }
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderWangsuCDNPro, {
  getInitialValues,
  getSchema,
});

export default _default;
