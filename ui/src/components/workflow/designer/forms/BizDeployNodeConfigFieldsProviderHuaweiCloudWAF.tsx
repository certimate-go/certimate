import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_CLOUDSERVER = "cloudserver" as const;
const DEPLOY_TARGET_PREMIUMHOST = "premiumhost" as const;
const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;

const BizDeployNodeConfigFieldsProviderHuaweiCloudWAF = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "deployTarget"], formInst);

  return (
    <>
      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.huaweicloud_waf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_waf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_waf_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_CLOUDSERVER, DEPLOY_TARGET_PREMIUMHOST, DEPLOY_TARGET_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.huaweicloud_waf_deploy_target.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_CLOUDSERVER || fieldResourceType === DEPLOY_TARGET_PREMIUMHOST}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.huaweicloud_waf_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_waf_domain.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === DEPLOY_TARGET_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_waf_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(),
      deployTarget: z.enum([DEPLOY_TARGET_CLOUDSERVER, DEPLOY_TARGET_PREMIUMHOST, DEPLOY_TARGET_CERTIFICATE]),
      certificateId: z.string().nullish(),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_CLOUDSERVER:
        case DEPLOY_TARGET_PREMIUMHOST:
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

        case DEPLOY_TARGET_CERTIFICATE:
          {
            const scCertificateId = z.string().nonempty();
            const spCertificateId = scCertificateId.safeParse(values.certificateId);
            if (!spCertificateId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spCertificateId.error).errors.join(),
                path: ["certificateId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderHuaweiCloudWAF, {
  getInitialValues,
  getSchema,
});

export default _default;
