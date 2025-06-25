import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";

type DeployNodeConfigFormTencentCloudGAAPConfigFieldValues = Nullish<{
  endpoint?: string;
  resourceType: string;
  proxyId?: string;
  listenerId?: string;
}>;

export type DeployNodeConfigFormTencentCloudGAAPConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormTencentCloudGAAPConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormTencentCloudGAAPConfigFieldValues) => void;
};

const RESOURCE_TYPE_LISTENER = "listener" as const;

const initFormModel = (): DeployNodeConfigFormTencentCloudGAAPConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_LISTENER,
    listenerId: "",
  };
};

const DeployNodeConfigFormTencentCloudGAAPConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormTencentCloudGAAPConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    endpoint: z.string().nullish(),
    resourceType: z.literal(RESOURCE_TYPE_LISTENER, t("workflow_node.deploy.form.tencentcloud_gaap_resource_type.placeholder")),
    proxyId: z.string().nullish(),
    listenerId: z
      .string()
      .nullish()
      .refine(
        (v) => ![RESOURCE_TYPE_LISTENER].includes(fieldResourceType) || !!v?.trim(),
        t("workflow_node.deploy.form.tencentcloud_gaap_listener_id.placeholder")
      ),
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
        label={t("workflow_node.deploy.form.tencentcloud_gaap_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_gaap_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_gaap_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.tencentcloud_gaap_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.tencentcloud_gaap_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_LISTENER} value={RESOURCE_TYPE_LISTENER}>
            {t("workflow_node.deploy.form.tencentcloud_gaap_resource_type.option.listener.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="proxyId"
        label={t("workflow_node.deploy.form.tencentcloud_gaap_proxy_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_gaap_proxy_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.tencentcloud_gaap_proxy_id.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="listenerId"
          label={t("workflow_node.deploy.form.tencentcloud_gaap_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_gaap_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_gaap_listener_id.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormTencentCloudGAAPConfig;
