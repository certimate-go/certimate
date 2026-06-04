import { getI18n, useTranslation } from "react-i18next";
import { AutoComplete, Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { matchSearchOption } from "@/utils/search";
import { isDomain, isPortNumber } from "@/utils/validator";

import { useFormNestedFieldsContext } from "./_context";

const SERVICE_TYPE_CLOUDRESOURCE = "cloudresource" as const;
const SERVICE_TYPE_CNAME = "cname" as const;

const BizDeployNodeConfigFieldsProviderAliyunWAF = () => {
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
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.aliyun_waf_region.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_waf_region.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_waf_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "serviceVersion"]}
        initialValue={initialValues.serviceVersion}
        label={t("workflow_node.deploy.form.aliyun_waf_service_version.label")}
        rules={[formRule]}
      >
        <Select options={["3.0"].map((s) => ({ label: s, value: s }))} placeholder={t("workflow_node.deploy.form.aliyun_waf_service_version.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "serviceType"]}
        initialValue={initialValues.serviceType}
        label={t("workflow_node.deploy.form.aliyun_waf_service_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[SERVICE_TYPE_CLOUDRESOURCE, SERVICE_TYPE_CNAME].map((s) => ({
            label: t(`workflow_node.deploy.form.aliyun_waf_service_type.option.${s}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.aliyun_waf_service_type.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "instanceId"]}
        initialValue={initialValues.instanceId}
        label={t("workflow_node.deploy.form.aliyun_waf_instance_id.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.aliyun_waf_instance_id.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.aliyun_waf_instance_id.placeholder")} />
      </Form.Item>

      <Show when={fieldServiceType === SERVICE_TYPE_CLOUDRESOURCE}>
        <Form.Item
          name={[parentNamePath, "resourceProduct"]}
          initialValue={initialValues.resourceProduct}
          label={t("workflow_node.deploy.form.aliyun_waf_resource_product.label")}
          rules={[formRule]}
        >
          <AutoComplete
            options={["ecs", "clb4", "clb7", "nlb"].map((value) => ({ value }))}
            placeholder={t("workflow_node.deploy.form.aliyun_waf_resource_product.placeholder")}
            showSearch={{
              filterOption: (inputValue, option) => matchSearchOption(inputValue, option!),
            }}
          />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "resourceId"]}
          initialValue={initialValues.resourceId}
          label={t("workflow_node.deploy.form.aliyun_waf_resource_id.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.aliyun_waf_resource_id.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "resourcePort"]}
          initialValue={initialValues.resourcePort}
          label={t("workflow_node.deploy.form.aliyun_waf_resource_port.label")}
          rules={[formRule]}
        >
          <Input type="number" min={1} max={65535} placeholder={t("workflow_node.deploy.form.aliyun_waf_resource_port.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "domain"]}
        initialValue={initialValues.domain}
        label={t("workflow_node.deploy.form.aliyun_waf_domain.label")}
        extra={t("workflow_node.deploy.form.aliyun_waf_domain.help")}
        rules={[formRule]}
      >
        <Input allowClear placeholder={t("workflow_node.deploy.form.aliyun_waf_domain.placeholder")} />
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    serviceVersion: "3.0",
    instanceId: "",
    resourceProduct: "",
    resourceId: "",
    resourcePort: 443,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(),
      serviceVersion: z.enum(["3.0"]),
      serviceType: z.enum([SERVICE_TYPE_CLOUDRESOURCE, SERVICE_TYPE_CNAME]),
      instanceId: z.string().nonempty(),
      resourceProduct: z.string().nullish(),
      resourceId: z.string().nullish(),
      resourcePort: z.union([z.string(), z.int().positive()]).nullish(),
      domain: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return isDomain(v, { allowWildcard: true });
        }, t("common.errmsg.domain_invalid")),
    })
    .superRefine((values, ctx) => {
      switch (values.serviceType) {
        case SERVICE_TYPE_CLOUDRESOURCE:
          {
            const scResourceProduct = z.string().nonempty();
            const spResourceProduct = scResourceProduct.safeParse(values.resourceProduct);
            if (!spResourceProduct.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spResourceProduct.error).errors.join(),
                path: ["resourceProduct"],
              });
            }

            const scResourceId = z.string().nonempty();
            const spResourceId = scResourceId.safeParse(values.resourceId);
            if (!spResourceId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spResourceId.error).errors.join(),
                path: ["resourceId"],
              });
            }

            const scResourcePort = z.coerce.number().refine((v) => isPortNumber(v), t("common.errmsg.port_invalid"));
            const spResourcePort = scResourcePort.safeParse(values.resourcePort);
            if (!spResourcePort.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spResourcePort.error).errors.join(),
                path: ["resourcePort"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderAliyunWAF, {
  getInitialValues,
  getSchema,
});

export default _default;
