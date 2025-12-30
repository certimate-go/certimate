import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_WEBSITE = "website" as const;
const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const WEBSITE_MATCH_PATTERN_SPECIFIED = "specified" as const;
const WEBSITE_MATCH_PATTERN_CERTSAN = "certsan" as const;

const BizDeployNodeConfigFieldsProvider1Panel = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "resourceType"], formInst);
  const fieldWebsiteMatchPattern = Form.useWatch([parentNamePath, "websiteMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "nodeName"]}
        initialValue={initialValues.nodeName}
        label={t("workflow_node.deploy.form.1panel_node_name.label")}
        extra={t("workflow_node.deploy.form.1panel_node_name.help")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_node_name.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.1panel_node_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_WEBSITE, RESOURCE_TYPE_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.1panel_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_WEBSITE}>
        <Form.Item
          name={[parentNamePath, "websiteMatchPattern"]}
          initialValue={initialValues.websiteMatchPattern}
          label={t("workflow_node.deploy.form.1panel_website_match_pattern.label")}
          extra={
            fieldWebsiteMatchPattern === WEBSITE_MATCH_PATTERN_CERTSAN ? (
              <span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_website_match_pattern.help_certsan") }}></span>
            ) : (
              void 0
            )
          }
          rules={[formRule]}
        >
          <Radio.Group
            options={[WEBSITE_MATCH_PATTERN_SPECIFIED, WEBSITE_MATCH_PATTERN_CERTSAN].map((s) => ({
              key: s,
              label: t(`workflow_node.deploy.form.1panel_website_match_pattern.option.${s}.label`),
              value: s,
            }))}
          />
        </Form.Item>

        <Show when={fieldWebsiteMatchPattern !== WEBSITE_MATCH_PATTERN_CERTSAN}>
          <Form.Item
            name={[parentNamePath, "websiteId"]}
            initialValue={initialValues.websiteId}
            label={t("workflow_node.deploy.form.1panel_website_id.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_website_id.tooltip") }}></span>}
          >
            <Input type="number" placeholder={t("workflow_node.deploy.form.1panel_website_id.placeholder")} />
          </Form.Item>
        </Show>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.1panel_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.1panel_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.1panel_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    resourceType: RESOURCE_TYPE_WEBSITE,
    websiteMatchPattern: WEBSITE_MATCH_PATTERN_SPECIFIED,
    websiteId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      nodeName: z.string().nullish(),
      resourceType: z.literal([RESOURCE_TYPE_WEBSITE, RESOURCE_TYPE_CERTIFICATE], t("workflow_node.deploy.form.shared_resource_type.placeholder")),
      websiteMatchPattern: z.string().nullish(),
      websiteId: z.union([z.string(), z.number().int()]).nullish(),
      certificateId: z.union([z.string(), z.number().int()]).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_WEBSITE:
          {
            if (values.websiteMatchPattern) {
              switch (values.websiteMatchPattern) {
                case WEBSITE_MATCH_PATTERN_SPECIFIED:
                  {
                    const scWebsiteId = z.coerce.number().int().positive();
                    if (!scWebsiteId.safeParse(values.websiteId).success) {
                      ctx.addIssue({
                        code: "custom",
                        message: t("workflow_node.deploy.form.1panel_website_id.placeholder"),
                        path: ["websiteId"],
                      });
                    }
                  }
                  break;
              }
            } else {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.1panel_website_match_pattern.placeholder"),
                path: ["websiteMatchPattern"],
              });
            }
          }
          break;

        case RESOURCE_TYPE_CERTIFICATE:
          {
            const scCertificateId = z.coerce.number().int().positive();
            if (!scCertificateId.safeParse(values.certificateId).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.1panel_certificate_id.placeholder"),
                path: ["websiteId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProvider1Panel, {
  getInitialValues,
  getSchema,
});

export default _default;
