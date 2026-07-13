import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;

const BizDeployNodeConfigFieldsProviderHuaweiCloudVOD = () => {
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
        label={t("workflow_node.deploy.form.huaweicloud_vod_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_vod_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_vod_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domainMatchPattern"]}
        initialValue={initialValues.domainMatchPattern}
        label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
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
        label={t("workflow_node.deploy.form.huaweicloud_vod_domain.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_vod_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(),
      domainMatchPattern: z.string().nonempty().default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern) {
        switch (values.domainMatchPattern) {
          case DOMAIN_MATCH_PATTERN_EXACT:
            {
              const scDomain = z.string().refine((v) => isDomain(v), t("common.errmsg.domain_invalid"));
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

const _default = Object.assign(BizDeployNodeConfigFieldsProviderHuaweiCloudVOD, {
  getInitialValues,
  getSchema,
});

export default _default;
