import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_VSERVER = "vserver" as const;

const BizDeployNodeConfigFieldsProviderUCloudUCLB = () => {
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
        label={t("workflow_node.deploy.form.ucloud_uclb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ucloud_uclb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ucloud_uclb_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_VSERVER].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.ucloud_uclb_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
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

      <Show when={fieldResourceType === RESOURCE_TYPE_VSERVER}>
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
    resourceType: RESOURCE_TYPE_VSERVER,
    loadbalancerId: "",
    vserverId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      endpoint: z.string().nullish(),
      resourceType: z.literal([RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_VSERVER], t("workflow_node.deploy.form.shared_resource_type.placeholder")),
      region: z.string().nonempty(t("workflow_node.deploy.form.ucloud_uclb_region.placeholder")),
      loadbalancerId: z.string().nonempty(t("workflow_node.deploy.form.ucloud_uclb_loadbalancer_id.placeholder")),
      vserverId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_VSERVER:
          {
            if (!values.vserverId?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ucloud_uclb_vserver_id.placeholder"),
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
