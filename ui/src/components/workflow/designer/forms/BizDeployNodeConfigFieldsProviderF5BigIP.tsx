import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;
const DEPLOY_TARGET_CLIENTSSL = "clientssl" as const;

const BizDeployNodeConfigFieldsProviderF5BigIP = () => {
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
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.f5bigip.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_CERTIFICATE, DEPLOY_TARGET_CLIENTSSL].map((s) => ({
            label: t(`workflow_node.deploy.form.f5bigip_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "partition"]}
        initialValue={initialValues.partition}
        label={t("workflow_node.deploy.form.f5bigip_partition.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.f5bigip_partition.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.f5bigip_partition.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_CLIENTSSL}>
        <Form.Item
          name={[parentNamePath, "clientSSLProfileName"]}
          initialValue={initialValues.clientSSLProfileName}
          label={t("workflow_node.deploy.form.f5bigip_client_ssl_profile_name.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.f5bigip_client_ssl_profile_name.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.f5bigip_client_ssl_profile_name.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    deployTarget: DEPLOY_TARGET_CERTIFICATE,
    partition: "Common",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_CERTIFICATE, DEPLOY_TARGET_CLIENTSSL]),
      partition: z.string().nonempty(),
      clientSSLProfileName: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_CLIENTSSL:
          {
            const scClientSSLProfileName = z.string().nonempty();
            const spClientSSLProfileName = scClientSSLProfileName.safeParse(values.clientSSLProfileName);
            if (!spClientSSLProfileName.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spClientSSLProfileName.error).errors.join(),
                path: ["clientSSLProfileName"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderF5BigIP, {
  getInitialValues,
  getSchema,
});

export default _default;
