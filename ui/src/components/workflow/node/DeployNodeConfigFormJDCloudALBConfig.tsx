import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormJDCloudALBConfigFieldValues = Nullish<{
  regionId: string;
  resourceType: string;
  loadbalancerId?: string;
  listenerId?: string;
  domain?: string;
}>;

export type DeployNodeConfigFormJDCloudALBConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormJDCloudALBConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormJDCloudALBConfigFieldValues) => void;
};

const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;

const initFormModel = (): DeployNodeConfigFormJDCloudALBConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_LISTENER,
  };
};

const DeployNodeConfigFormJDCloudALBConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormJDCloudALBConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    regionId: z
      .string(t("workflow_node.deploy.form.jdcloud_alb_region_id.placeholder"))
      .nonempty(t("workflow_node.deploy.form.jdcloud_alb_region_id.placeholder")),
    resourceType: z.literal([RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER], t("workflow_node.deploy.form.jdcloud_alb_resource_type.placeholder")),
    loadbalancerId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => fieldResourceType !== RESOURCE_TYPE_LOADBALANCER || !!v?.trim(), t("workflow_node.deploy.form.jdcloud_alb_loadbalancer_id.placeholder")),
    listenerId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => fieldResourceType !== RESOURCE_TYPE_LISTENER || !!v?.trim(), t("workflow_node.deploy.form.jdcloud_alb_listener_id.placeholder")),
    domain: z
      .string()
      .nullish()
      .refine((v) => {
        if (![RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER].includes(fieldResourceType)) return true;
        return !v || validDomainName(v!, { allowWildcard: true });
      }, t("common.errmsg.domain_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const fieldResourceType = Form.useWatch("resourceType", formInst);

  const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
    onValuesChange?.(values);
  };

  return (
    <Form
      form={formInst}
      disabled={disabled}
      initialValues={initialValues ?? initFormModel()}
      layout="vertical"
      name={formName}
      onValuesChange={handleFormChange}
    >
      <Form.Item
        name="regionId"
        label={t("workflow_node.deploy.form.jdcloud_alb_region_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_alb_region_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.jdcloud_alb_region_id.placeholder")} />
      </Form.Item>

      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.jdcloud_alb_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.jdcloud_alb_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_LOADBALANCER} value={RESOURCE_TYPE_LOADBALANCER}>
            {t("workflow_node.deploy.form.jdcloud_alb_resource_type.option.loadbalancer.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_LISTENER} value={RESOURCE_TYPE_LISTENER}>
            {t("workflow_node.deploy.form.jdcloud_alb_resource_type.option.listener.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_LOADBALANCER}>
        <Form.Item
          name="loadbalancerId"
          label={t("workflow_node.deploy.form.jdcloud_alb_loadbalancer_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_alb_loadbalancer_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.jdcloud_alb_loadbalancer_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="listenerId"
          label={t("workflow_node.deploy.form.jdcloud_alb_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_alb_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.jdcloud_alb_listener_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LOADBALANCER || fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="domain"
          label={t("workflow_node.deploy.form.jdcloud_alb_snidomain.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.jdcloud_alb_snidomain.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.jdcloud_alb_snidomain.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormJDCloudALBConfig;
