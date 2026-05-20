import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_WEBSITE = "website" as const;
const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;

const WEBSITE_MATCH_PATTERN_SPECIFIED = "specified" as const;
const WEBSITE_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProvider1Panel = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "deployTarget"], formInst);
  const fieldWebsiteMatchPattern = Form.useWatch([parentNamePath, "websiteMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "nodeName"]}
        initialValue={initialValues.nodeName}
        label={t("workflow_node.deploy.form.1panel_node_name.label")}
        extra={t("workflow_node.deploy.form.1panel_node_name.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_node_name.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.1panel_node_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_WEBSITE, DEPLOY_TARGET_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.1panel_deploy_target.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_WEBSITE}>
        <Form.Item
          name={[parentNamePath, "websiteMatchPattern"]}
          initialValue={initialValues.websiteMatchPattern}
          label={t("workflow_node.deploy.form.1panel_website_match_pattern.label")}
          extra={
            fieldWebsiteMatchPattern === WEBSITE_MATCH_PATTERN_CERTSAN ? (
              <span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_website_match_pattern.option.certsan.help") }}></span>
            ) : (
              void 0
            )
          }
          rules={[formRule]}
        >
          <Radio.Group
            options={[WEBSITE_MATCH_PATTERN_SPECIFIED, WEBSITE_MATCH_PATTERN_CERTSAN].map((s) => ({
              key: s,
              label: t(`workflow_node.deploy.form.1panel_website_match_pattern.option.${s}.label`),
              value: s,
            }))}
          />
        </Form.Item>

        <Show when={fieldWebsiteMatchPattern !== WEBSITE_MATCH_PATTERN_CERTSAN}>
          <Form.Item
            name={[parentNamePath, "websiteId"]}
            initialValue={initialValues.websiteId}
            label={t("workflow_node.deploy.form.1panel_website_id.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_website_id.tooltip") }}></span>}
          >
            <Input type="number" placeholder={t("workflow_node.deploy.form.1panel_website_id.placeholder")} />
          </Form.Item>
        </Show>
      </Show>

      <Show when={fieldResourceType === DEPLOY_TARGET_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.1panel_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.1panel_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_WEBSITE,
    websiteMatchPattern: WEBSITE_MATCH_PATTERN_SPECIFIED,
    websiteId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      nodeName: z.string().nullish(),
      deployTarget: z.enum([DEPLOY_TARGET_WEBSITE, DEPLOY_TARGET_CERTIFICATE]),
      websiteMatchPattern: z.string().nullish().default(WEBSITE_MATCH_PATTERN_SPECIFIED),
      websiteId: z.union([z.string(), z.int().positive()]).nullish(),
      certificateId: z.union([z.string(), z.int().positive()]).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_WEBSITE:
          {
            const scWebsiteMatchPattern = z.string().nonempty();
            const spWebsiteMatchPattern = scWebsiteMatchPattern.safeParse(values.websiteMatchPattern);
            if (!spWebsiteMatchPattern.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spWebsiteMatchPattern.error).errors.join(),
                path: ["websiteMatchPattern"],
              });
            }

            switch (values.websiteMatchPattern) {
              case WEBSITE_MATCH_PATTERN_SPECIFIED:
                {
                  const scWebsiteId = z.coerce.number().int().positive();
                  const spWebsiteId = scWebsiteId.safeParse(values.websiteId);
                  if (!spWebsiteId.success) {
                    ctx.addIssue({
                      code: "custom",
                      message: z.treeifyError(spWebsiteId.error).errors.join(),
                      path: ["websiteId"],
                    });
                  }
                }
                break;
            }
          }
          break;

        case DEPLOY_TARGET_CERTIFICATE:
          {
            const scCertificateId = z.coerce.number().int().positive();
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

const _default = Object.assign(BizDeployNodeConfigFieldsProvider1Panel, {
  getInitialValues,
  getSchema,
});

export default _default;
