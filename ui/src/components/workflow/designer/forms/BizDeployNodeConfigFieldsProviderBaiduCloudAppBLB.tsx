import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain, isPortNumber } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_LOADBALANCER = "loadbalancer" as const;
const DEPLOY_TARGET_LISTENER = "listener" as const;

const BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB = () => {
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
        label={t("workflow_node.deploy.form.baiducloud_appblb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baiducloud_appblb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_LOADBALANCER, DEPLOY_TARGET_LISTENER].map((s) => ({
            label: t(`workflow_node.deploy.form.baiducloud_appblb_deploy_target.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "loadbalancerId"]}
        initialValue={initialValues.loadbalancerId}
        label={t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_LISTENER}>
        <Form.Item
          name={[parentNamePath, "listenerPort"]}
          initialValue={initialValues.listenerPort}
          label={t("workflow_node.deploy.form.baiducloud_appblb_listener_port.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_listener_port.tooltip") }}></span>}
        >
          <Input type="number" min={1} max={65535} placeholder={t("workflow_node.deploy.form.baiducloud_appblb_listener_port.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === DEPLOY_TARGET_LOADBALANCER || fieldResourceType === DEPLOY_TARGET_LISTENER}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.baiducloud_appblb_snidomain.label")}
          extra={t("workflow_node.deploy.form.baiducloud_appblb_snidomain.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.baiducloud_appblb_snidomain.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    deployTarget: DEPLOY_TARGET_LISTENER,
    loadbalancerId: "",
    listenerPort: 443,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(),
      deployTarget: z.enum([DEPLOY_TARGET_LOADBALANCER, DEPLOY_TARGET_LISTENER]),
      loadbalancerId: z.string().nonempty(),
      listenerPort: z.union([z.string(), z.int().positive()]).nullish(),
      domain: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return isDomain(v, { allowWildcard: true });
        }, t("common.errmsg.domain_invalid")),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_LISTENER:
          {
            const scListenerPort = z.coerce.number().refine((v) => isPortNumber(v), t("common.errmsg.port_invalid"));
            const spListenerPort = scListenerPort.safeParse(values.listenerPort);
            if (!spListenerPort.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spListenerPort.error).errors.join(),
                path: ["listenerPort"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB, {
  getInitialValues,
  getSchema,
});

export default _default;
