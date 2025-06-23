import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";

type DeployNodeConfigFormHuaweiCloudELBConfigFieldValues = Nullish<{
  region: string;
  resourceType: string;
  certificateId?: string;
  loadbalancerId?: string;
  listenerId?: string;
}>;

export type DeployNodeConfigFormHuaweiCloudELBConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormHuaweiCloudELBConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormHuaweiCloudELBConfigFieldValues) => void;
};

const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;
const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;

const initFormModel = (): DeployNodeConfigFormHuaweiCloudELBConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_LISTENER,
  };
};

const DeployNodeConfigFormHuaweiCloudELBConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormHuaweiCloudELBConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.huaweicloud_elb_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.huaweicloud_elb_region.placeholder")),
    resourceType: z.literal(
      [RESOURCE_TYPE_CERTIFICATE, RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER],
      t("workflow_node.deploy.form.huaweicloud_elb_resource_type.placeholder")
    ),
    certificateId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => fieldResourceType !== RESOURCE_TYPE_CERTIFICATE || !!v?.trim(), t("workflow_node.deploy.form.huaweicloud_elb_certificate_id.placeholder")),
    loadbalancerId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine(
        (v) => fieldResourceType !== RESOURCE_TYPE_LOADBALANCER || !!v?.trim(),
        t("workflow_node.deploy.form.huaweicloud_elb_loadbalancer_id.placeholder")
      ),
    listenerId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => fieldResourceType !== RESOURCE_TYPE_LISTENER || !!v?.trim(), t("workflow_node.deploy.form.huaweicloud_elb_listener_id.placeholder")),
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
        name="region"
        label={t("workflow_node.deploy.form.huaweicloud_elb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_elb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.huaweicloud_elb_region.placeholder")} />
      </Form.Item>

      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.huaweicloud_elb_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.huaweicloud_elb_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_CERTIFICATE} value={RESOURCE_TYPE_CERTIFICATE}>
            {t("workflow_node.deploy.form.huaweicloud_elb_resource_type.option.certificate.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_LOADBALANCER} value={RESOURCE_TYPE_LOADBALANCER}>
            {t("workflow_node.deploy.form.huaweicloud_elb_resource_type.option.loadbalancer.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_LISTENER} value={RESOURCE_TYPE_LISTENER}>
            {t("workflow_node.deploy.form.huaweicloud_elb_resource_type.option.listener.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name="certificateId"
          label={t("workflow_node.deploy.form.huaweicloud_elb_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_elb_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_elb_certificate_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LOADBALANCER}>
        <Form.Item
          name="loadbalancerId"
          label={t("workflow_node.deploy.form.huaweicloud_elb_loadbalancer_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_elb_loadbalancer_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_elb_loadbalancer_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="listenerId"
          label={t("workflow_node.deploy.form.huaweicloud_elb_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.huaweicloud_elb_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.huaweicloud_elb_listener_id.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormHuaweiCloudELBConfig;
