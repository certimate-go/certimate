import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormRainYunRCDNConfigFieldValues = Nullish<{
  instanceId: string | number;
  domain: string;
}>;

export type DeployNodeConfigFormRainYunRCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormRainYunRCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormRainYunRCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormRainYunRCDNConfigFieldValues => {
  return {
    instanceId: "",
  };
};

const DeployNodeConfigFormRainYunRCDNConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormRainYunRCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    instanceId: z.union([z.string(), z.number().int()]).refine((v) => {
      return /^\d+$/.test(v + "") && +v > 0;
    }, t("workflow_node.deploy.form.rainyun_rcdn_instance_id.placeholder")),
    domain: z
      .string(t("workflow_node.deploy.form.rainyun_rcdn_domain.placeholder"))
      .refine((v) => validDomainName(v, { allowWildcard: true }), t("common.errmsg.domain_invalid")),
  });
  const formRule = createSchemaFieldRule(formSchema);

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
        name="instanceId"
        label={t("workflow_node.deploy.form.rainyun_rcdn_instance_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.rainyun_rcdn_instance_id.tooltip") }}></span>}
      >
        <Input type="number" placeholder={t("workflow_node.deploy.form.rainyun_rcdn_instance_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="domain"
        label={t("workflow_node.deploy.form.rainyun_rcdn_domain.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.rainyun_rcdn_domain.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.rainyun_rcdn_domain.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormRainYunRCDNConfig;
