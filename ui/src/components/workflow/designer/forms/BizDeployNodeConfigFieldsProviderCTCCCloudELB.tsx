import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_LOADBALANCER = "loadbalancer" as const;
const DEPLOY_TARGET_LISTENER = "listener" as const;

const BizDeployNodeConfigFieldsProviderCTCCCloudELB = () => {
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
        name={[parentNamePath, "regionId"]}
        initialValue={initialValues.regionId}
        label={t("workflow_node.deploy.form.ctcccloud_elb_region_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ctcccloud_elb_region_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ctcccloud_elb_region_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_LOADBALANCER, DEPLOY_TARGET_LISTENER].map((s) => ({
            label: t(`workflow_node.deploy.form.ctcccloud_elb_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_LOADBALANCER}>
        <Form.Item
          name={[parentNamePath, "loadbalancerId"]}
          initialValue={initialValues.loadbalancerId}
          label={t("workflow_node.deploy.form.ctcccloud_elb_loadbalancer_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ctcccloud_elb_loadbalancer_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ctcccloud_elb_loadbalancer_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === DEPLOY_TARGET_LISTENER}>
        <Form.Item
          name={[parentNamePath, "listenerId"]}
          initialValue={initialValues.listenerId}
          label={t("workflow_node.deploy.form.ctcccloud_elb_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ctcccloud_elb_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ctcccloud_elb_listener_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    regionId: "",
    deployTarget: DEPLOY_TARGET_LISTENER,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      regionId: z.string().nonempty(),
      deployTarget: z.enum([DEPLOY_TARGET_LOADBALANCER, DEPLOY_TARGET_LISTENER]),
      loadbalancerId: z.string().nullish(),
      listenerId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_LOADBALANCER:
          {
            const scLoadbalancerId = z.string().nonempty();
            const spLoadbalancerId = scLoadbalancerId.safeParse(values.loadbalancerId);
            if (!spLoadbalancerId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spLoadbalancerId.error).errors.join(),
                path: ["loadbalancerId"],
              });
            }
          }
          break;

        case DEPLOY_TARGET_LISTENER:
          {
            const scListenerId = z.string().nonempty();
            const spListenerId = scListenerId.safeParse(values.listenerId);
            if (!spListenerId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spListenerId.error).errors.join(),
                path: ["listenerId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderCTCCCloudELB, {
  getInitialValues,
  getSchema,
});

export default _default;
