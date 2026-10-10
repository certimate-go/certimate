import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;
const DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard" as const;
const DOMAIN_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProviderAliyunMSE = () => {
  const { i18n, t } = useTranslation();
  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({ [parentNamePath]: getSchema({ i18n }) });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();
  const fieldDomainMatchPattern = Form.useWatch([parentNamePath, "domainMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.aliyun_mse_region.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_mse_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "gatewayId"]}
        initialValue={initialValues.gatewayId}
        label={t("workflow_node.deploy.form.aliyun_mse_gateway_id.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_mse_gateway_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domainMatchPattern"]}
        initialValue={initialValues.domainMatchPattern}
        label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
        extra={
          <>
            <div>{t("workflow_node.deploy.form.aliyun_mse_domain.help")}</div>
            {fieldDomainMatchPattern === DOMAIN_MATCH_PATTERN_EXACT && (
              <span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_domain_match_pattern.option.exact.help.wildcard") }}></span>
            )}
          </>
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
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.aliyun_mse_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_mse_domain.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "forceHttps"]}
        initialValue={initialValues.forceHttps}
        label={t("workflow_node.deploy.form.aliyun_mse_force_https.label")}
        extra={t("workflow_node.deploy.form.aliyun_mse_force_https.help")}
        rules={[formRule]}
      >
        <Switch />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    gatewayId: "",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domain: "",
    forceHttps: false,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().trim().nonempty(),
      gatewayId: z.string().trim().nonempty(),
      domainMatchPattern: z.enum([DOMAIN_MATCH_PATTERN_EXACT, DOMAIN_MATCH_PATTERN_WILDCARD, DOMAIN_MATCH_PATTERN_CERTSAN]).default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().trim().nullish(),
      forceHttps: z.boolean().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern !== DOMAIN_MATCH_PATTERN_CERTSAN && !isDomain(values.domain ?? "", { allowWildcard: true })) {
        ctx.addIssue({
          code: "custom",
          message: t("common.errmsg.domain_invalid"),
          path: ["domain"],
        });
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAliyunMSE, {
  getInitialValues,
  getSchema,
});

export default _default;
