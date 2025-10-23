import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const SITE_TYPE_PHP = "php";
const SITE_TYPE_OTHER = "other";

const MULTIPLE_INPUT_SEPARATOR = ";";

const BizDeployNodeConfigFieldsProviderBaotaPanelSite = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldSiteType = Form.useWatch([parentNamePath, "siteType"], formInst);

  return (
    <>
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanel_site.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "siteType"]}
        initialValue={initialValues.siteType}
        label={t("workflow_node.deploy.form.baotapanel_site_type.label")}
        rules={[formRule]}
      >
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
          name={[parentNamePath, "siteName"]}
          initialValue={initialValues.siteName}
          label={t("workflow_node.deploy.form.baotapanel_site_name.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.baotapanel_site_name.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.baotapanel_site_name.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldSiteType === SITE_TYPE_OTHER}>
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
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    siteType: SITE_TYPE_OTHER,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      siteType: z.literal([SITE_TYPE_PHP, SITE_TYPE_OTHER], t("workflow_node.deploy.form.baotapanel_site_type.placeholder")),
      siteName: z.string().nullish(),
      siteNames: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.siteType) {
        case SITE_TYPE_PHP:
          {
            if (!values.siteName?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.baotapanel_site_name.placeholder"),
                path: ["siteName"],
              });
            }
          }
          break;

        case SITE_TYPE_OTHER:
          {
            if (!values.siteNames?.trim() || !values.siteNames.split(MULTIPLE_INPUT_SEPARATOR).every((e) => !!e.trim())) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.baotapanel_site_names.placeholder"),
                path: ["siteNames"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderBaotaPanelSite, {
  getInitialValues,
  getSchema,
});

export default _default;
