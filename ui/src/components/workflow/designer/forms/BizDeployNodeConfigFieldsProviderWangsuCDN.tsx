import { getI18n, useTranslation } from "react-i18next";
import { Form, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;

const BizDeployNodeConfigFieldsProviderWangsuCDN = () => {
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
            key: s,
            label: t(`workflow_node.deploy.form.shared_domain_match_pattern.option.${s}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domains"]}
        initialValue={initialValues.domains}
        label={t("workflow_node.deploy.form.wangsu_cdn_domains.label")}
        extra={t("workflow_node.deploy.form.wangsu_cdn_domains.help")}
        rules={[formRule]}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.wangsu_cdn_domains.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.wangsu_cdn_domains.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.wangsu_cdn_domains.multiple_input_modal.placeholder")}
          separator={MULTIPLE_INPUT_SEPARATOR}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domains: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      domainMatchPattern: z.string().nonempty(t("workflow_node.deploy.form.shared_domain_match_pattern.placeholder")).default(DOMAIN_MATCH_PATTERN_EXACT),
      domains: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern) {
        switch (values.domainMatchPattern) {
          case DOMAIN_MATCH_PATTERN_EXACT:
            {
              const valid = values.domains && values.domains.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isDomain(e, { allowWildcard: true }));
              if (!valid) {
                ctx.addIssue({
                  code: "custom",
                  message: t("common.errmsg.domain_invalid"),
                  path: ["domains"],
                });
              }
            }
            break;
        }
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderWangsuCDN, {
  getInitialValues,
  getSchema,
});

export default _default;
