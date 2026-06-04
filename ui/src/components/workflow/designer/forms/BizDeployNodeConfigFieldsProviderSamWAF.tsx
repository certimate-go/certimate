import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;

const BizDeployNodeConfigFieldsProviderSamWAF = () => {
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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.samwaf.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_CERTIFICATE].map((s) => ({
            label: t(`workflow_node.deploy.form.samwaf_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.samwaf_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.samwaf_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.samwaf_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_CERTIFICATE,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_CERTIFICATE]),
      certificateId: z.union([z.string(), z.int().positive()]).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
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

const _default = Object.assign(BizDeployNodeConfigFieldsProviderSamWAF, {
  getInitialValues,
  getSchema,
});

export default _default;
