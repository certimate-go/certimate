import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";

type DeployNodeConfigFormBaotaPanelSiteConfigFieldValues = Nullish<{
  siteType: string;
  siteName?: string;
  siteNames?: string;
}>;

export type DeployNodeConfigFormBaotaPanelSiteConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormBaotaPanelSiteConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormBaotaPanelSiteConfigFieldValues) => void;
};

const SITE_TYPE_PHP = "php";
const SITE_TYPE_OTHER = "other";

const MULTIPLE_INPUT_SEPARATOR = ";";

const initFormModel = (): DeployNodeConfigFormBaotaPanelSiteConfigFieldValues => {
  return {
    siteType: SITE_TYPE_OTHER,
    siteName: "",
    siteNames: "",
  };
};

const DeployNodeConfigFormBaotaPanelSiteConfig = ({
  form: formInst,
  formName,
  disabled,
  initialValues,
  onValuesChange,
}: DeployNodeConfigFormBaotaPanelSiteConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    siteType: z.literal([SITE_TYPE_PHP, SITE_TYPE_OTHER], t("workflow_node.deploy.form.baotapanel_site_type.placeholder")),
    siteName: z
      .string()
      .nullish()
      .refine((v) => {
        if (fieldSiteType !== SITE_TYPE_PHP) return true;
        return !!v?.trim();
      }, t("workflow_node.deploy.form.baotapanel_site_name.placeholder")),
    siteNames: z
      .string()
      .nullish()
      .refine((v) => {
        if (fieldSiteType !== SITE_TYPE_OTHER) return true;
        if (!v) return false;
        return String(v)
          .split(MULTIPLE_INPUT_SEPARATOR)
          .every((e) => !!e.trim());
      }, t("workflow_node.deploy.form.baotapanel_site_names.placeholder")),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const fieldSiteType = Form.useWatch<string>("siteType", formInst);

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
      <Form.Item name="siteType" label={t("workflow_node.deploy.form.baotapanel_site_type.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.baotapanel_site_type.placeholder")}>
          <Select.Option key={SITE_TYPE_PHP} value={SITE_TYPE_PHP}>
            {t("workflow_node.deploy.form.baotapanel_site_type.option.php.label")}
          </Select.Option>
          <Select.Option key={SITE_TYPE_OTHER} value={SITE_TYPE_OTHER}>
            {t("workflow_node.deploy.form.baotapanel_site_type.option.other.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Show when={fieldSiteType === SITE_TYPE_PHP}>
        <Form.Item
          name="siteName"
          label={t("workflow_node.deploy.form.baotapanel_site_name.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanel_site_name.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.baotapanel_site_name.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldSiteType === SITE_TYPE_OTHER}>
        <Form.Item
          name="siteNames"
          label={t("workflow_node.deploy.form.baotapanel_site_names.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanel_site_names.tooltip") }}></span>}
        >
          <MultipleSplitValueInput
            modalTitle={t("workflow_node.deploy.form.baotapanel_site_names.multiple_input_modal.title")}
            placeholder={t("workflow_node.deploy.form.baotapanel_site_names.placeholder")}
            placeholderInModal={t("workflow_node.deploy.form.baotapanel_site_names.multiple_input_modal.placeholder")}
            splitOptions={{ trim: true, removeEmpty: true }}
          />
        </Form.Item>
      </Show>
    </Form>
  );
};

export default DeployNodeConfigFormBaotaPanelSiteConfig;
