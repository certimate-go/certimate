import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormTencentCloudCLBConfigFieldValues = Nullish<{
  endpoint?: string;
  region: string;
  resourceType: string;
  loadbalancerId?: string;
  listenerId?: string;
  domain?: string;
}>;

export type DeployNodeConfigFormTencentCloudCLBConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudCLBConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudCLBConfigFieldValues) => void;
};

const RESOURCE_TYPE_LOADBALANCER = "loadbalancer" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;
const RESOURCE_TYPE_RULEDOMAIN = "ruledomain" as const;
const RESOURCE_TYPE_VIA_SSLDEPLOY = "ssl-deploy" as const;

const initFormModel = (): DeployNodeConfigFormTencentCloudCLBConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_LISTENER,
  };
};

const DeployNodeConfigFormTencentCloudCLBConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudCLBConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    resourceType: z.literal(
      [RESOURCE_TYPE_LOADBALANCER, RESOURCE_TYPE_LISTENER, RESOURCE_TYPE_RULEDOMAIN, RESOURCE_TYPE_VIA_SSLDEPLOY],
      t("workflow_node.deploy.form.tencentcloud_clb_resource_type.placeholder")
    ),
    region: z
      .string(t("workflow_node.deploy.form.tencentcloud_clb_region.placeholder"))
      .nonempty(t("workflow_node.deploy.form.tencentcloud_clb_region.placeholder")),
    loadbalancerId: z
      .string()
      .min(1, t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    listenerId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine(
        (v) => ![RESOURCE_TYPE_LISTENER, RESOURCE_TYPE_RULEDOMAIN, RESOURCE_TYPE_VIA_SSLDEPLOY].includes(fieldResourceType) || !!v?.trim(),
        t("workflow_node.deploy.form.tencentcloud_clb_listener_id.placeholder")
      ),
    domain: z
      .string()
      .nullish()
      .refine((v) => RESOURCE_TYPE_RULEDOMAIN !== fieldResourceType || validDomainName(v!, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
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
        name="endpoint"
        label={t("workflow_node.deploy.form.tencentcloud_clb_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_clb_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name="region"
        label={t("workflow_node.deploy.form.tencentcloud_clb_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_region.placeholder")} />
      </Form.Item>

      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.tencentcloud_clb_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.tencentcloud_clb_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_LOADBALANCER} value={RESOURCE_TYPE_LOADBALANCER}>
            {t("workflow_node.deploy.form.tencentcloud_clb_resource_type.option.loadbalancer.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_LISTENER} value={RESOURCE_TYPE_LISTENER}>
            {t("workflow_node.deploy.form.tencentcloud_clb_resource_type.option.listener.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_RULEDOMAIN} value={RESOURCE_TYPE_RULEDOMAIN}>
            {t("workflow_node.deploy.form.tencentcloud_clb_resource_type.option.ruledomain.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_VIA_SSLDEPLOY} value={RESOURCE_TYPE_VIA_SSLDEPLOY}>
            {t("workflow_node.deploy.form.tencentcloud_clb_resource_type.option.ssl_deploy.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="loadbalancerId"
        label={t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_loadbalancer_id.placeholder")} />
      </Form.Item>

      <Show
        when={
          fieldResourceType === RESOURCE_TYPE_VIA_SSLDEPLOY || fieldResourceType === RESOURCE_TYPE_LISTENER || fieldResourceType === RESOURCE_TYPE_RULEDOMAIN
        }
      >
        <Form.Item
          name="listenerId"
          label={t("workflow_node.deploy.form.tencentcloud_clb_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_listener_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_VIA_SSLDEPLOY}>
        <Form.Item
          name="domain"
          label={t("workflow_node.deploy.form.tencentcloud_clb_snidomain.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_snidomain.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_clb_snidomain.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_RULEDOMAIN}>
        <Form.Item
          name="domain"
          label={t("workflow_node.deploy.form.tencentcloud_clb_ruledomain.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_clb_ruledomain.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_clb_ruledomain.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudCLBConfig;
