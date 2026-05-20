import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;
const RESOURCE_TYPE_RULEDOMAIN = "ruledomain" as const;

const BizDeployNodeConfigFieldsProviderTencentCloudCLB = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "resourceType"], formInst);

  return (
    <>
      <Form.Item
        name={[parentNamePath, "endpoint"]}
        initialValue={initialValues.endpoint}
        label={t("workflow_node.deploy.form.tencentcloud_clb_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_clb_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.tencentcloud_clb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER, RESOURCE_TYPE_RULEDOMAIN].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.tencentcloud_clb_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "loadbalancerId"]}
        initialValue={initialValues.loadbalancerId}
        label={t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER || fieldResourceType === RESOURCE_TYPE_RULEDOMAIN}>
        <Form.Item
          name={[parentNamePath, "listenerId"]}
          initialValue={initialValues.listenerId}
          label={t("workflow_node.deploy.form.tencentcloud_clb_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_listener_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_RULEDOMAIN}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.tencentcloud_clb_ruledomain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_ruledomain.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    resourceType: RESOURCE_TYPE_LISTENER,
    loadbalancerId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      endpoint: z.string().nullish(),
      resourceType: z.enum([RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER, RESOURCE_TYPE_RULEDOMAIN]),
      region: z.string().nonempty(),
      loadbalancerId: z.string().nonempty(),
      listenerId: z.string().nullish(),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_LISTENER:
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

        case RESOURCE_TYPE_RULEDOMAIN:
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

            const scDomain = z.string().refine((v) => isDomain(v, { allowWildcard: true }), t("common.errmsg.domain_invalid"));
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

const _default = Object.assign(BizDeployNodeConfigFieldsProviderTencentCloudCLB, {
  getInitialValues,
  getSchema,
});

export default _default;
