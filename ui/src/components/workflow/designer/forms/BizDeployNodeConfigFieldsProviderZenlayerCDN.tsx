import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_DOMAIN = "domain" as const;
const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;
const DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard" as const;
const DOMAIN_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProviderZenlayerCDN = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "resourceType"], formInst);
  const fieldDomainMatchPattern = Form.useWatch([parentNamePath, "domainMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_DOMAIN, RESOURCE_TYPE_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.zenlayer_cdn_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_DOMAIN}>
        <Form.Item
          name={[parentNamePath, "domainMatchPattern"]}
          initialValue={initialValues.domainMatchPattern}
          label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
          extra={
            fieldDomainMatchPattern === DOMAIN_MATCH_PATTERN_EXACT ? (
              <span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_domain_match_pattern.help_wildcard") }}></span>
            ) : (
              void 0
            )
          }
          rules={[formRule]}
        >
          <Radio.Group
            options={[DOMAIN_MATCH_PATTERN_EXACT, DOMAIN_MATCH_PATTERN_WILDCARD, DOMAIN_MATCH_PATTERN_CERTSAN].map((s) => ({
              key: s,
              label: t(`workflow_node.deploy.form.shared_domain_match_pattern.option.${s}.label`),
              value: s,
            }))}
          />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.zenlayer_cdn_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.zenlayer_cdn_domain.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.zenlayer_cdn_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.zenlayer_cdn_certificate_id.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.zenlayer_cdn_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    resourceType: RESOURCE_TYPE_DOMAIN,
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      resourceType: z.literal([RESOURCE_TYPE_DOMAIN, RESOURCE_TYPE_CERTIFICATE], t("workflow_node.deploy.form.shared_resource_type.placeholder")),
      domainMatchPattern: z.string().nonempty(t("workflow_node.deploy.form.shared_domain_match_pattern.placeholder")).default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().nullish(),
      certificateId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_DOMAIN:
          {
            if (!values.domainMatchPattern) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.shared_domain_match_pattern.placeholder"),
                path: ["domainMatchPattern"],
              });
            }

            switch (values.domainMatchPattern) {
              case DOMAIN_MATCH_PATTERN_EXACT:
              case DOMAIN_MATCH_PATTERN_WILDCARD:
                {
                  if (!isDomain(values.domain!, { allowWildcard: true })) {
                    ctx.addIssue({
                      code: "custom",
                      message: t("common.errmsg.domain_invalid"),
                      path: ["domain"],
                    });
                  }
                }
                break;
            }
          }
          break;

        case RESOURCE_TYPE_CERTIFICATE:
          {
            if (!values.certificateId) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.zenlayer_cdn_certificate_id.placeholder"),
                path: ["certificateId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderZenlayerCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
