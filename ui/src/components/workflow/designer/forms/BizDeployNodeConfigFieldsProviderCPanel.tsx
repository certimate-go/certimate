import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_WEBSITE = "website" as const;

const BizDeployNodeConfigFieldsProviderCPanel = () => {
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
            value: s,
            label: t(`workflow_node.deploy.form.cpanel_deploy_target.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_WEBSITE}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.cpanel_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.cpanel_domain.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_WEBSITE,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_WEBSITE]),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_WEBSITE:
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
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderCPanel, {
  getInitialValues,
  getSchema,
});

export default _default;
