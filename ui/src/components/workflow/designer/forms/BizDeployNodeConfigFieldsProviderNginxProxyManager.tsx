import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const RESOURCE_TYPE_HOST = "host" as const;
const RESOURCE_TYPE_CERTIFICATE = "certificate" as const;

const HOST_MATCH_PATTERN_SPECIFIED = "specified" as const;
const HOST_MATCH_PATTERN_CERTSAN = "certsan" as const;

const HOST_TYPE_PROXY = "proxy" as const;
const HOST_TYPE_REDIRECTION = "redirection" as const;
const HOST_TYPE_STREAM = "stream" as const;
const HOST_TYPE_DEAD = "dead" as const;

const BizDeployNodeConfigFieldsProviderNginxProxyManager = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldResourceType = Form.useWatch([parentNamePath, "resourceType"], formInst);
  const fieldHostMatchPattern = Form.useWatch([parentNamePath, "hostMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "resourceType"]}
        initialValue={initialValues.resourceType}
        label={t("workflow_node.deploy.form.shared_resource_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[RESOURCE_TYPE_HOST, RESOURCE_TYPE_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.nginxproxymanager_resource_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_resource_type.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === RESOURCE_TYPE_HOST}>
        <Form.Item
          name={[parentNamePath, "hostMatchPattern"]}
          initialValue={initialValues.hostMatchPattern}
          label={t("workflow_node.deploy.form.nginxproxymanager_host_match_pattern.label")}
          rules={[formRule]}
        >
          <Radio.Group
            options={[HOST_MATCH_PATTERN_SPECIFIED, HOST_MATCH_PATTERN_CERTSAN].map((s) => ({
              key: s,
              label: t(`workflow_node.deploy.form.nginxproxymanager_host_match_pattern.option.${s}.label`),
              value: s,
            }))}
          />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "hostType"]}
          initialValue={initialValues.hostId}
          label={t("workflow_node.deploy.form.nginxproxymanager_host_type.label")}
          rules={[formRule]}
        >
          <Select
            options={[HOST_TYPE_PROXY, HOST_TYPE_REDIRECTION, HOST_TYPE_STREAM, HOST_TYPE_DEAD].map((s) => ({
              value: s,
              label: t(`workflow_node.deploy.form.nginxproxymanager_host_type.option.${s}.label`),
            }))}
            placeholder={t("workflow_node.deploy.form.nginxproxymanager_host_type.placeholder")}
          />
        </Form.Item>

        <Show when={fieldHostMatchPattern !== HOST_MATCH_PATTERN_CERTSAN}>
          <Form.Item
            name={[parentNamePath, "hostId"]}
            initialValue={initialValues.hostId}
            label={t("workflow_node.deploy.form.nginxproxymanager_host_id.label")}
            rules={[formRule]}
            tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.nginxproxymanager_host_id.tooltip") }}></span>}
          >
            <Input type="number" placeholder={t("workflow_node.deploy.form.nginxproxymanager_host_id.placeholder")} />
          </Form.Item>
        </Show>
      </Show>

      <Show when={fieldResourceType === RESOURCE_TYPE_CERTIFICATE}>
        <Form.Item
          name={[parentNamePath, "certificateId"]}
          initialValue={initialValues.certificateId}
          label={t("workflow_node.deploy.form.nginxproxymanager_certificate_id.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.nginxproxymanager_certificate_id.tooltip") }}></span>}
        >
          <Input type="number" placeholder={t("workflow_node.deploy.form.nginxproxymanager_certificate_id.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    resourceType: RESOURCE_TYPE_HOST,
    hostMatchPattern: HOST_MATCH_PATTERN_SPECIFIED,
    hostType: HOST_TYPE_PROXY,
    hostId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      resourceType: z.literal([RESOURCE_TYPE_HOST, RESOURCE_TYPE_CERTIFICATE], t("workflow_node.deploy.form.shared_resource_type.placeholder")),
      hostMatchPattern: z.string().nullish(),
      hostType: z.string().nullish(),
      hostId: z.union([z.string(), z.number().int()]).nullish(),
      certificateId: z.union([z.string(), z.number().int()]).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.resourceType) {
        case RESOURCE_TYPE_HOST:
          {
            if (values.hostMatchPattern) {
              switch (values.hostMatchPattern) {
                case HOST_MATCH_PATTERN_SPECIFIED:
                  {
                    const scHostType = z.string().nonempty();
                    if (!scHostType.safeParse(values.hostType).success) {
                      ctx.addIssue({
                        code: "custom",
                        message: t("workflow_node.deploy.form.nginxproxymanager_host_type.placeholder"),
                        path: ["hostType"],
                      });
                    }

                    const scHostId = z.coerce.number().int().positive();
                    if (!scHostId.safeParse(values.hostId).success) {
                      ctx.addIssue({
                        code: "custom",
                        message: t("workflow_node.deploy.form.nginxproxymanager_host_id.placeholder"),
                        path: ["hostId"],
                      });
                    }
                  }
                  break;
              }
            } else {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.nginxproxymanager_host_match_pattern.placeholder"),
                path: ["hostMatchPattern"],
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
                message: t("workflow_node.deploy.form.nginxproxymanager_certificate_id.placeholder"),
                path: ["hostId"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderNginxProxyManager, {
  getInitialValues,
  getSchema,
});

export default _default;
