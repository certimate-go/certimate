import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_WEBSITE = "website" as const;

const BizDeployNodeConfigFieldsProviderNetlify = () => {
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
          options={[DEPLOY_TARGET_WEBSITE].map((s) => ({
            label: t(`workflow_node.deploy.form.netlify_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_WEBSITE}>
        <Form.Item
          name={[parentNamePath, "siteId"]}
          initialValue={initialValues.siteId}
          label={t("workflow_node.deploy.form.netlify_site_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.netlify_site_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.netlify_site_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_WEBSITE,
    siteId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_WEBSITE]),
      siteId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_WEBSITE:
          {
            const scSiteId = z.string().nonempty();
            const spSiteId = scSiteId.safeParse(values.siteId);
            if (!spSiteId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spSiteId.error).errors.join(),
                path: ["siteId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderNetlify, {
  getInitialValues,
  getSchema,
});

export default _default;
