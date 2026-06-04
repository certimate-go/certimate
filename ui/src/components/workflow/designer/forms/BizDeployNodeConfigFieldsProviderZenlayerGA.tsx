import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_ACCELERATOR = "accelerator" as const;
const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;

const BizDeployNodeConfigFieldsProviderZenlayerGA = () => {
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
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_ACCELERATOR, DEPLOY_TARGET_CERTIFICATE].map((s) => ({
            label: t(`workflow_node.deploy.form.zenlayer_ga_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_ACCELERATOR}>
        <Form.Item
          name={[parentNamePath, "acceleratorId"]}
          initialValue={initialValues.acceleratorId}
          label={t("workflow_node.deploy.form.zenlayer_ga_accelerator_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.zenlayer_ga_accelerator_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.zenlayer_ga_accelerator_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === DEPLOY_TARGET_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.zenlayer_ga_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.zenlayer_ga_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.zenlayer_ga_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_ACCELERATOR,
    acceleratorId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_ACCELERATOR, DEPLOY_TARGET_CERTIFICATE]),
      acceleratorId: z.string().nullish(),
      certificateId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_ACCELERATOR:
          {
            const scAcceleratorId = z.string().nonempty();
            const spAcceleratorId = scAcceleratorId.safeParse(values.acceleratorId);
            if (!spAcceleratorId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spAcceleratorId.error).errors.join(),
                path: ["acceleratorId"],
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

const _default = Object.assign(BizDeployNodeConfigFieldsProviderZenlayerGA, {
  getInitialValues,
  getSchema,
});

export default _default;
