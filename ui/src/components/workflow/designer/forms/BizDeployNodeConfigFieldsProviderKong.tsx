import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import Tips from "@/components/Tips";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const BizDeployNodeConfigFieldsProviderKong = () => {
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
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.kong.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.kong_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "workspace"]}
        initialValue={initialValues.workspace}
        label={t("workflow_node.deploy.form.kong_workspace.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.kong_workspace.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.kong_workspace.placeholder")} />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.kong_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.kong_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.kong_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    resourceType: RESOURCE_TYPE_CERTIFICATE,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      resourceType: z.literal(RESOURCE_TYPE_CERTIFICATE, t("workflow_node.deploy.form.shared_resource_type.placeholder")),
      workspace: z.string().nullish(),
      certificateId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_CERTIFICATE:
          {
            if (!values.certificateId?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.kong_certificate_id.placeholder"),
                path: ["certificateId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderKong, {
  getInitialValues,
  getSchema,
});

export default _default;
