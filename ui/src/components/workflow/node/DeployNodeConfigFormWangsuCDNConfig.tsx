import { useTranslation } from "react-i18next";
import { Form, type FormInstance } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import { validDomainName } from "@/utils/validators";

type DeployNodeConfigFormWangsuCDNConfigFieldValues = Nullish<{
  domains: string;
}>;

export type DeployNodeConfigFormWangsuCDNConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormWangsuCDNConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormWangsuCDNConfigFieldValues) => void;
};

const MULTIPLE_INPUT_SEPARATOR = ";";

const initFormModel = (): DeployNodeConfigFormWangsuCDNConfigFieldValues => {
  return {
    domains: "",
  };
};

const DeployNodeConfigFormWangsuCDNConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormWangsuCDNConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    domains: z
      .string()
      .nullish()
      .refine((v) => {
        if (!v) return false;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => validDomainName(e, { allowWildcard: true }));
      }, t("workflow_node.deploy.form.wangsu_cdn_domains.placeholder")),
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
        name="domains"
        label={t("workflow_node.deploy.form.wangsu_cdn_domains.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.wangsu_cdn_domains.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.wangsu_cdn_domains.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.wangsu_cdn_domains.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.wangsu_cdn_domains.multiple_input_modal.placeholder")}
          splitOptions={{ trim: true, removeEmpty: true }}
        />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormWangsuCDNConfig;
