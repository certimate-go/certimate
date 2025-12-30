import { getI18n, useTranslation } from "react-i18next";
import { Form, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderBaotaPanel = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const initialValues = getInitialValues();

  return (
    <>
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanel.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "siteType"]}
        initialValue={initialValues.siteType}
        label={t("workflow_node.deploy.form.baotapanel_site_type.label")}
        rules={[formRule]}
      >
        <Select
          options={["php", "any"].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.baotapanel_site_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.baotapanel_site_type.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "siteNames"]}
        initialValue={initialValues.siteNames}
        label={t("workflow_node.deploy.form.baotapanel_site_names.label")}
        extra={t("workflow_node.deploy.form.baotapanel_site_names.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanel_site_names.tooltip") }}></span>}
      >
        <MultipleSplitValueInput
          modalTitle={t("workflow_node.deploy.form.baotapanel_site_names.multiple_input_modal.title")}
          placeholder={t("workflow_node.deploy.form.baotapanel_site_names.placeholder")}
          placeholderInModal={t("workflow_node.deploy.form.baotapanel_site_names.multiple_input_modal.placeholder")}
          separator={MULTIPLE_INPUT_SEPARATOR}
          splitOptions={{ removeEmpty: true, trimSpace: true }}
        />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    siteType: "any",
    siteNames: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z.object({
    siteType: z.literal(["php", "any"], t("workflow_node.deploy.form.baotapanel_site_type.placeholder")),
    siteNames: z
      .string()
      .nonempty(t("workflow_node.deploy.form.baotapanel_site_names.placeholder"))
      .refine(
        (v) => {
          if (!v) return false;
          return v.split(MULTIPLE_INPUT_SEPARATOR).every((s) => !!s.trim());
        },
        { error: t("workflow_node.deploy.form.baotapanel_site_names.placeholder") }
      ),
  });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaotaPanel, {
  getInitialValues,
  getSchema,
});

export default _default;
