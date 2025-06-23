import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import Show from "@/components/Show";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormAliyunGAConfigFieldValues = Nullish<{
  resourceType: string;
  acceleratorId?: string;
  listenerId?: string;
  domain?: string;
}>;

export type DeployNodeConfigFormAliyunGAConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormAliyunGAConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormAliyunGAConfigFieldValues) => void;
};

const RESOURCE_TYPE_ACCELERATOR = "accelerator" as const;
const RESOURCE_TYPE_LISTENER = "listener" as const;

const initFormModel = (): DeployNodeConfigFormAliyunGAConfigFieldValues => {
  return {
    resourceType: RESOURCE_TYPE_LISTENER,
  };
};

const DeployNodeConfigFormAliyunGAConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormAliyunGAConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    resourceType: z.literal([RESOURCE_TYPE_ACCELERATOR, RESOURCE_TYPE_LISTENER], t("workflow_node.deploy.form.aliyun_ga_resource_type.placeholder")),
    acceleratorId: z.string().max(64, t("common.errmsg.string_max", { max: 64 })),
    listenerId: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 64 }))
      .nullish()
      .refine((v) => fieldResourceType !== RESOURCE_TYPE_LISTENER || !!v?.trim(), t("workflow_node.deploy.form.aliyun_ga_listener_id.placeholder")),
    domain: z
      .string()
      .nullish()
      .refine((v) => {
        if (![RESOURCE_TYPE_ACCELERATOR, RESOURCE_TYPE_LISTENER].includes(fieldResourceType)) return true;
        return !v || validDomainName(v!);
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
      <Form.Item name="resourceType" label={t("workflow_node.deploy.form.aliyun_ga_resource_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.aliyun_ga_resource_type.placeholder")}>
          <Select.Option key={RESOURCE_TYPE_ACCELERATOR} value={RESOURCE_TYPE_ACCELERATOR}>
            {t("workflow_node.deploy.form.aliyun_ga_resource_type.option.accelerator.label")}
          </Select.Option>
          <Select.Option key={RESOURCE_TYPE_LISTENER} value={RESOURCE_TYPE_LISTENER}>
            {t("workflow_node.deploy.form.aliyun_ga_resource_type.option.listener.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="acceleratorId"
        label={t("workflow_node.deploy.form.aliyun_ga_accelerator_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_ga_accelerator_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_ga_accelerator_id.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="listenerId"
          label={t("workflow_node.deploy.form.aliyun_ga_listener_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_ga_listener_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_ga_listener_id.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_ACCELERATOR || fieldResourceType === RESOURCE_TYPE_LISTENER}>
        <Form.Item
          name="domain"
          label={t("workflow_node.deploy.form.aliyun_ga_snidomain.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_ga_snidomain.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.aliyun_ga_snidomain.placeholder")} />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormAliyunGAConfig;
