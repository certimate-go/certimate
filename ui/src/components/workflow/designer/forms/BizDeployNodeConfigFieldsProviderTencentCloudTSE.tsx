import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import MultipleSplitValueInput from "@/components/MultipleSplitValueInput";
import Show from "@/components/Show";
import { isDomain } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const MULTIPLE_INPUT_SEPARATOR = ";";

const SERVICE_TYPE_CLOUDNATIVE = "cloudnative" as const;

const BizDeployNodeConfigFieldsProviderTencentCloudTSE = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldServiceType = Form.useWatch([parentNamePath, "serviceType"], formInst);

  return (
    <>
      <Form.Item
        name={[parentNamePath, "endpoint"]}
        initialValue={initialValues.endpoint}
        label={t("workflow_node.deploy.form.tencentcloud_tse_endpoint.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_tse_endpoint.tooltip") }}></span>}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.tencentcloud_tse_endpoint.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "serviceType"]}
        initialValue={initialValues.serviceType}
        label={t("workflow_node.deploy.form.tencentcloud_tse_service_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[SERVICE_TYPE_CLOUDNATIVE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.tencentcloud_tse_service_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.tencentcloud_tse_service_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldServiceType === SERVICE_TYPE_CLOUDNATIVE}>
        <Form.Item
          name={[parentNamePath, "gatewayId"]}
          initialValue={initialValues.gatewayId}
          label={t("workflow_node.deploy.form.tencentcloud_tse_gateway_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_tse_gateway_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_tse_gateway_id.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "domains"]}
          initialValue={initialValues.domains}
          label={t("workflow_node.deploy.form.tencentcloud_tse_domains.label")}
          extra={t("workflow_node.deploy.form.tencentcloud_tse_domains.help")}
          rules={[formRule]}
        >
          <MultipleSplitValueInput
            modalTitle={t("workflow_node.deploy.form.tencentcloud_tse_domains.multiple_input_modal.title")}
            placeholder={t("workflow_node.deploy.form.tencentcloud_tse_domains.placeholder")}
            placeholderInModal={t("workflow_node.deploy.form.tencentcloud_tse_domains.multiple_input_modal.placeholder")}
            splitOptions={{ removeEmpty: true, trimSpace: true }}
          />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.tencentcloud_tse_certificate_id.label")}
          extra={t("workflow_node.deploy.form.tencentcloud_tse_certificate_id.help")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.tencentcloud_tse_certificate_id.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.tencentcloud_tse_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    serviceType: SERVICE_TYPE_CLOUDNATIVE,
    gatewayId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      serviceType: z.enum([SERVICE_TYPE_CLOUDNATIVE]),
      endpoint: z.string().nullish(),
      region: z.string().nonempty(),
      gatewayId: z.string().nullish(),
      domains: z.string().nullish(),
      certificateId: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.serviceType) {
        switch (values.serviceType) {
          case SERVICE_TYPE_CLOUDNATIVE:
            {
              const scGatewayId = z.string().nonempty();
              const spGatewayId = scGatewayId.safeParse(values.gatewayId);
              if (!spGatewayId.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spGatewayId.error).errors.join(),
                  path: ["gatewayId"],
                });
              }

              const scDomains = z
                .string()
                .nonempty()
                .refine((v) => v.split(MULTIPLE_INPUT_SEPARATOR).every((e) => isDomain(e, { allowWildcard: true })), t("common.errmsg.domain_invalid"));
              const spDomains = scDomains.safeParse(values.domains);
              if (!spDomains.success) {
                ctx.addIssue({
                  code: "custom",
                  message: z.treeifyError(spDomains.error).errors.join(),
                  path: ["domains"],
                });
              }
            }
            break;
        }
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderTencentCloudTSE, {
  getInitialValues,
  getSchema,
});

export default _default;
