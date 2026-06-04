import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_WEBSITE = "website" as const;
const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderRatPanel = () => {
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
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ratpanel.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_WEBSITE, DEPLOY_TARGET_CERTIFICATE].map((s) => ({
            label: t(`workflow_node.deploy.form.ratpanel_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_WEBSITE}>
        <Form.Item
          name={[parentNamePath, "siteNames"]}
          initialValue={initialValues.siteNames}
          label={t("workflow_node.deploy.form.ratpanel_site_names.label")}
          extra={t("workflow_node.deploy.form.ratpanel_site_names.help")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ratpanel_site_names.tooltip") }}></span>}
        >
          <MultipleSplitValueInput
            modalTitle={t("workflow_node.deploy.form.ratpanel_site_names.multiple_input_modal.title")}
            placeholder={t("workflow_node.deploy.form.ratpanel_site_names.placeholder")}
            placeholderInModal={t("workflow_node.deploy.form.ratpanel_site_names.multiple_input_modal.placeholder")}
            separator={MULTIPLE_INPUT_SEPARATOR}
            splitOptions={{ removeEmpty: true, trimSpace: true }}
          />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === DEPLOY_TARGET_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.ratpanel_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ratpanel_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.ratpanel_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_WEBSITE,
    siteNames: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_WEBSITE, DEPLOY_TARGET_CERTIFICATE]),
      siteNames: z.string().nullish(),
      certificateId: z.union([z.string(), z.int().positive()]).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_WEBSITE:
          {
            const scSiteNames = z
              .string()
              .nonempty()
              .refine((v) => {
                return v.split(MULTIPLE_INPUT_SEPARATOR).every((s) => !!s.trim());
              });
            const spSiteNames = scSiteNames.safeParse(values.siteNames);
            if (!spSiteNames.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spSiteNames.error).errors.join(),
                path: ["siteNames"],
              });
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

const _default = Object.assign(BizDeployNodeConfigFieldsProviderRatPanel, {
  getInitialValues,
  getSchema,
});

export default _default;
