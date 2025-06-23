import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { validDomainName, validPortNumber } from "@/utils/validators";

type DeployNodeConfigFormBaiduCloudAppBLBConfigFieldValues = Nullish<{
  region: string;
  resourceType: string;
  loadbalancerId?: string;
  listenerPort?: number;
  domain?: string;
}>;

export type DeployNodeConfigFormBaiduCloudAppBLBConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormBaiduCloudAppBLBConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormBaiduCloudAppBLBConfigFieldValues) => void;
};

const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;

const initFormModel = (): DeployNodeConfigFormBaiduCloudAppBLBConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_LISTENER,
    listenerPort: 443,
  };
};

const DeployNodeConfigFormBaiduCloudAppBLBConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormBaiduCloudAppBLBConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    region: z
      .string(t("workflow_node.deploy.form.baiducloud_appblb_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.baiducloud_appblb_region.placeholder")),
    resourceType: z.literal([RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER], t("workflow_node.deploy.form.baiducloud_appblb_resource_type.placeholder")),
    loadbalancerId: z
      .string(t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.placeholder"))
      .min(1, t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    listenerPort: z.preprocess(
      (v) => (v == null || v === "" ? undefined : Number(v)),
      z
        .number()
        .refine(
          (v) => fieldResourceType === RESOURCE_TYPE_LISTENER && validPortNumber(v!),
          t("workflow_node.deploy.form.baiducloud_appblb_listener_port.placeholder")
        )
        .nullish()
    ),
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
        name="region"
        label={t("workflow_node.deploy.form.baiducloud_appblb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baiducloud_appblb_region.placeholder")} />
      </Form.Item>

      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.baiducloud_appblb_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.baiducloud_appblb_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_LOADBALANCER} value={RESOURCE_TYPE_LOADBALANCER}>
            {t("workflow_node.deploy.form.baiducloud_appblb_resource_type.option.loadbalancer.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_LISTENER} value={RESOURCE_TYPE_LISTENER}>
            {t("workflow_node.deploy.form.baiducloud_appblb_resource_type.option.listener.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="loadbalancerId"
        label={t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baiducloud_appblb_loadbalancer_id.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="listenerPort"
          label={t("workflow_node.deploy.form.baiducloud_appblb_listener_port.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_listener_port.tooltip") }}></span>}
        >
          <Input type="number" min={1} max={65535} placeholder={t("workflow_node.deploy.form.baiducloud_appblb_listener_port.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_LOADBALANCER || fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="domain"
          label={t("workflow_node.deploy.form.baiducloud_appblb_snidomain.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baiducloud_appblb_snidomain.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.baiducloud_appblb_snidomain.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormBaiduCloudAppBLBConfig;
