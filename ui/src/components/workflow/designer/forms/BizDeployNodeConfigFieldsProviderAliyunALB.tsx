import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;

const BizDeployNodeConfigFieldsProviderAliyunALB = () => {
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
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.aliyun_alb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_alb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_alb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.aliyun_alb_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_LOADBALANCER}>
        <Form.Item
          name={[parentNamePath, "loadbalancerId"]}
          initialValue={initialValues.loadbalancerId}
          label={t("workflow_node.deploy.form.aliyun_alb_loadbalancer_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_alb_loadbalancer_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_alb_loadbalancer_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name={[parentNamePath, "listenerId"]}
          initialValue={initialValues.listenerId}
          label={t("workflow_node.deploy.form.aliyun_alb_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_alb_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_alb_listener_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LOADBALANCER || fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.aliyun_alb_snidomain.label")}
          extra={t("workflow_node.deploy.form.aliyun_alb_snidomain.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.aliyun_alb_snidomain.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    resourceType: RESOURCE_TYPE_LISTENER,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(t("workflow_node.deploy.form.aliyun_alb_region.placeholder")),
      resourceType: z.literal([RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER], t("workflow_node.deploy.form.shared_resource_type.placeholder")),
      loadbalancerId: z.string().nullish(),
      listenerId: z.string().nullish(),
      domain: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return isDomain(v, { allowWildcard: true });
        }, t("common.errmsg.domain_invalid")),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_LOADBALANCER:
          {
            const scLoadbalancerId = z.string().nonempty();
            if (!scLoadbalancerId.safeParse(values.loadbalancerId).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.aliyun_alb_loadbalancer_id.placeholder"),
                path: ["loadbalancerId"],
              });
            }
          }
          break;

        case RESOURCE_TYPE_LISTENER:
          {
            const scListenerId = z.string().nonempty();
            if (!scListenerId.safeParse(values.listenerId).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.aliyun_alb_listener_id.placeholder"),
                path: ["listenerId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAliyunALB, {
  getInitialValues,
  getSchema,
});

export default _default;
