import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, InputNumber } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validPortNumber } from "@/utils/validators";

type DeployNodeConfigFormBaotaWAFSiteConfigFieldValues = Nullish<{
  siteName: string;
  sitePort: number;
}>;

export type DeployNodeConfigFormBaotaWAFSiteConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormBaotaWAFSiteConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormBaotaWAFSiteConfigFieldValues) => void;
};

const initFormModel = (): DeployNodeConfigFormBaotaWAFSiteConfigFieldValues => {
  return {
    siteName: "",
    sitePort: 443,
  };
};

const DeployNodeConfigFormBaotaWAFSiteConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormBaotaWAFSiteConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    siteName: z.string().nonempty(t("workflow_node.deploy.form.baotawaf_site_name.placeholder")),
    sitePort: z.preprocess(
      (v) => Number(v),
      z
        .number()
        .int(t("workflow_node.deploy.form.baotawaf_site_port.placeholder"))
        .refine((v) => validPortNumber(v), t("common.errmsg.port_invalid"))
    ),
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
        name="siteName"
        label={t("workflow_node.deploy.form.baotawaf_site_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotawaf_site_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.baotawaf_site_name.placeholder")} />
      </Form.Item>

      <Form.Item name="sitePort" label={t("workflow_node.deploy.form.baotawaf_site_port.label")} rules={[formRule]}>
        <InputNumber className="w-full" placeholder={t("access.form.ssh_port.placeholder")} min={1} max={65535} />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormBaotaWAFSiteConfig;
