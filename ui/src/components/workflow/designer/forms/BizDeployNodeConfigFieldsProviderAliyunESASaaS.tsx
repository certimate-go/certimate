import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;
const DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard" as const;
const DOMAIN_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProviderAliyunESASaaS = () => {
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
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.aliyun_esa_saas_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_esa_saas_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_esa_saas_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "siteId"]}
        initialValue={initialValues.siteId}
        label={t("workflow_node.deploy.form.aliyun_esa_saas_site_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_esa_saas_site_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("workflow_node.deploy.form.aliyun_esa_saas_site_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domainMatchPattern"]}
        initialValue={initialValues.domainMatchPattern}
        label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
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

      <Show when={fieldDomainMatchPattern !== DOMAIN_MATCH_PATTERN_CERTSAN}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.aliyun_esa_saas_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_esa_saas_domain.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    siteId: "",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(t("workflow_node.deploy.form.aliyun_esa_saas_region.placeholder")),
      siteId: z.union([z.string(), z.number().int()]).refine((v) => {
        return /^\d+$/.test(v + "") && +v > 0;
      }, t("workflow_node.deploy.form.aliyun_esa_saas_site_id.placeholder")),
      domainMatchPattern: z.string().nonempty(t("workflow_node.deploy.form.shared_domain_match_pattern.placeholder")).default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern) {
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
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAliyunESASaaS, {
  getInitialValues,
  getSchema,
});

export default _default;
