import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const SERVICE_TYPE_CLOUDNATIVE = "cloudnative" as const;
const SERVICE_TYPE_TRADITIONAL = "traditional" as const;

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;
const DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard" as const;
const DOMAIN_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProviderAliyunAPIGW = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldServiceType = Form.useWatch([parentNamePath, "serviceType"], formInst);
  const fieldDomainMatchPattern = Form.useWatch([parentNamePath, "domainMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.aliyun_apigw_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "serviceType"]}
        initialValue={initialValues.serviceType}
        label={t("workflow_node.deploy.form.aliyun_apigw_service_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[SERVICE_TYPE_CLOUDNATIVE, SERVICE_TYPE_CLOUDNATIVE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.aliyun_apigw_service_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.aliyun_apigw_service_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldServiceType === SERVICE_TYPE_CLOUDNATIVE}>
        <Form.Item
          name={[parentNamePath, "gatewayId"]}
          initialValue={initialValues.gatewayId}
          label={t("workflow_node.deploy.form.aliyun_apigw_gateway_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_gateway_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_gateway_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldServiceType === SERVICE_TYPE_TRADITIONAL}>
        <Form.Item
          name={[parentNamePath, "groupId"]}
          initialValue={initialValues.groupId}
          label={t("workflow_node.deploy.form.aliyun_apigw_group_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_apigw_group_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_group_id.placeholder")} />
        </Form.Item>
      </Show>

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
          label={t("workflow_node.deploy.form.aliyun_apigw_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_apigw_domain.placeholder")} />
        </Form.Item>
      </Show>
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
      serviceType: z.enum([SERVICE_TYPE_CLOUDNATIVE, SERVICE_TYPE_TRADITIONAL]),
      region: z.string().nonempty(),
      gatewayId: z.string().nullish(),
      groupId: z.string().nullish(),
      domainMatchPattern: z.string().nonempty().default(DOMAIN_MATCH_PATTERN_EXACT),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.serviceType) {
        switch (values.serviceType) {
          case SERVICE_TYPE_CLOUDNATIVE:
            {
              const scGatewayId = z.string().nonempty();
              const spGatewayId = scGatewayId.safeParse(values.gatewayId);
              if (!spGatewayId.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spGatewayId.error).errors.join(),
                  path: ["gatewayId"],
                });
              }
            }
            break;

          case SERVICE_TYPE_TRADITIONAL:
            {
              const scGroupId = z.string().nonempty();
              const spGroupId = scGroupId.safeParse(values.groupId);
              if (!spGroupId.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spGroupId.error).errors.join(),
                  path: ["groupId"],
                });
              }
            }
            break;
        }
      }

      if (values.domainMatchPattern) {
        switch (values.domainMatchPattern) {
          case DOMAIN_MATCH_PATTERN_EXACT:
          case DOMAIN_MATCH_PATTERN_WILDCARD:
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

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAliyunAPIGW, {
  getInitialValues,
  getSchema,
});

export default _default;
