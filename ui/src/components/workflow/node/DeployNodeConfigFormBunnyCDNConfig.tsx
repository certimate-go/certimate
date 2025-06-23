import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormBunnyCDNConfigFieldValues = Nullish<{
  pullZoneId: string | number;
  hostname: string;
}>;

export type DeployNodeConfigFormBunnyCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormBunnyCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormBunnyCDNConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormBunnyCDNConfigFieldValues => {
  return {};
};

const DeployNodeConfigFormBunnyCDNConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormBunnyCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    pullZoneId: z.union([z.string(), z.number().int()]).refine((v) => {
      return /^\d+$/.test(v + "") && +v! > 0;
    }, t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.placeholder")),
    hostname: z
      .string(t("workflow_node.deploy.form.bunny_cdn_hostname.placeholder"))
      .nonempty(t("workflow_node.deploy.form.bunny_cdn_hostname.placeholder"))
      .refine((v) => {
        return validDomainName(v!, { allowWildcard: true });
      }, t("common.errmsg.domain_invalid")),
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
        name="pullZoneId"
        label={t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.bunny_cdn_pull_zone_id.placeholder")} />
      </Form.Item>

      <Form.Item
        name="hostname"
        label={t("workflow_node.deploy.form.bunny_cdn_hostname.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.bunny_cdn_hostname.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.bunny_cdn_hostname.placeholder")} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormBunnyCDNConfig;
