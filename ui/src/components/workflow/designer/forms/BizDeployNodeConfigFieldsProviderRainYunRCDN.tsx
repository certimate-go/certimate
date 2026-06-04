import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;

const BizDeployNodeConfigFieldsProviderRainYunRCDN = () => {
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
        name={[parentNamePath, "instanceId"]}
        initialValue={initialValues.instanceId}
        label={t("workflow_node.deploy.form.rainyun_rcdn_instance_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.rainyun_rcdn_instance_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("workflow_node.deploy.form.rainyun_rcdn_instance_id.placeholder")} />
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
        label={t("workflow_node.deploy.form.rainyun_rcdn_domain.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.rainyun_rcdn_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    instanceId: "",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      instanceId: z.union([z.string(), z.int().positive()]).nullish(),
      domainMatchPattern: z.string().nonempty().default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
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
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderRainYunRCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
