import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_LOADBALANCER = "loadbalancer" as const;
const DEPLOY_TARGET_VSERVER = "vserver" as const;

const BizDeployNodeConfigFieldsProviderUCloudUCLB = () => {
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
        label={t("workflow_node.deploy.form.ucloud_uclb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_uclb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_uclb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_LOADBALANCER, DEPLOY_TARGET_VSERVER].map((s) => ({
            label: t(`workflow_node.deploy.form.ucloud_uclb_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "loadbalancerId"]}
        initialValue={initialValues.loadbalancerId}
        label={t("workflow_node.deploy.form.ucloud_uclb_loadbalancer_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_uclb_loadbalancer_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_uclb_loadbalancer_id.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_VSERVER}>
        <Form.Item
          name={[parentNamePath, "vserverId"]}
          initialValue={initialValues.vserverId}
          label={t("workflow_node.deploy.form.ucloud_uclb_vserver_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_uclb_vserver_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ucloud_uclb_vserver_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    deployTarget: DEPLOY_TARGET_VSERVER,
    loadbalancerId: "",
    vserverId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      endpoint: z.string().nullish(),
      deployTarget: z.enum([DEPLOY_TARGET_LOADBALANCER, DEPLOY_TARGET_VSERVER]),
      region: z.string().nonempty(),
      loadbalancerId: z.string().nonempty(),
      vserverId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_VSERVER:
          {
            const scVserverId = z.string().nonempty();
            const spVserverId = scVserverId.safeParse(values.vserverId);
            if (!spVserverId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spVserverId.error).errors.join(),
                path: ["vserverId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderUCloudUCLB, {
  getInitialValues,
  getSchema,
});

export default _default;
