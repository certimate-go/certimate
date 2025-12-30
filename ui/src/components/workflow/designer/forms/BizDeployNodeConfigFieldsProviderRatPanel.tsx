import { getI18n, useTranslation } from "react-i18next";
import { Form, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_WEBSITE = "website" as const;

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderRatPanel = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "resourceType"], formInst);

  return (
    <>
      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_WEBSITE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.ratpanel_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_WEBSITE}>
        <Form.Item
          name={[parentNamePath, "siteNames"]}
          initialValue={initialValues.siteNames}
          label={t("workflow_node.deploy.form.ratpanel_site_names.label")}
          extra={t("workflow_node.deploy.form.ratpanel_site_names.help")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ratpanel_site_names.tooltip") }}></span>}
        >
          <MultipleSplitValueInput
            modalTitle={t("workflow_node.deploy.form.ratpanel_site_names.multiple_input_modal.title")}
            placeholder={t("workflow_node.deploy.form.ratpanel_site_names.placeholder")}
            placeholderInModal={t("workflow_node.deploy.form.ratpanel_site_names.multiple_input_modal.placeholder")}
            separator={MULTIPLE_INPUT_SEPARATOR}
            splitOptions={{ removeEmpty: true, trimSpace: true }}
          />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    resourceType: RESOURCE_TYPE_WEBSITE,
    siteNames: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      resourceType: z.literal(RESOURCE_TYPE_WEBSITE, t("workflow_node.deploy.form.cpanel_resource_type.placeholder")),
      siteNames: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_WEBSITE:
          {
            const scSiteNames = z
              .string()
              .nonempty()
              .refine((v) => {
                if (!v) return false;
                return v.split(MULTIPLE_INPUT_SEPARATOR).every((s) => !!s.trim());
              });
            if (!scSiteNames.safeParse(values.siteNames).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ratpanel_site_names.placeholder"),
                path: ["siteNames"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderRatPanel, {
  getInitialValues,
  getSchema,
});

export default _default;
