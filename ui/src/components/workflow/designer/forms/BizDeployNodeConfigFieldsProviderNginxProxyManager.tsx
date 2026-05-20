import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";

import { useFormNestedFieldsContext } from "./_context";

const DEPLOY_TARGET_HOST = "host" as const;
const DEPLOY_TARGET_CERTIFICATE = "certificate" as const;

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

  const fieldResourceType = Form.useWatch([parentNamePath, "deployTarget"], formInst);
  const fieldHostMatchPattern = Form.useWatch([parentNamePath, "hostMatchPattern"], { form: formInst, preserve: true });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "deployTarget"]}
        initialValue={initialValues.deployTarget}
        label={t("workflow_node.deploy.form.shared_deploy_target.label")}
        rules={[formRule]}
      >
        <Select
          options={[DEPLOY_TARGET_HOST, DEPLOY_TARGET_CERTIFICATE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.nginxproxymanager_deploy_target.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.shared_deploy_target.placeholder")}
        />
      </Form.Item>

      <Show when={fieldResourceType === DEPLOY_TARGET_HOST}>
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

      <Show when={fieldResourceType === DEPLOY_TARGET_CERTIFICATE}>
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
    deployTarget: DEPLOY_TARGET_HOST,
    hostMatchPattern: HOST_MATCH_PATTERN_SPECIFIED,
    hostType: HOST_TYPE_PROXY,
    hostId: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      deployTarget: z.enum([DEPLOY_TARGET_HOST, DEPLOY_TARGET_CERTIFICATE]),
      hostMatchPattern: z.string().nullish(),
      hostType: z.string().nullish(),
      hostId: z.union([z.string(), z.int().positive()]).nullish(),
      certificateId: z.union([z.string(), z.int().positive()]).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.deployTarget) {
        case DEPLOY_TARGET_HOST:
          {
            const scHostMatchPattern = z.coerce.number().int().positive();
            const spHostMatchPattern = scHostMatchPattern.safeParse(values.hostMatchPattern);
            if (!spHostMatchPattern.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spHostMatchPattern.error).errors.join(),
                path: ["hostMatchPattern"],
              });
            }

            switch (values.hostMatchPattern) {
              case HOST_MATCH_PATTERN_SPECIFIED:
                {
                  const scHostType = z.string().nonempty();
                  const spHostType = scHostType.safeParse(values.hostType);
                  if (!spHostType.success) {
                    ctx.addIssue({
                      code: "custom",
                      message: z.treeifyError(spHostType.error).errors.join(),
                      path: ["hostType"],
                    });
                  }

                  const scHostId = z.coerce.number().int().positive();
                  const spHostId = scHostId.safeParse(values.hostId);
                  if (!spHostId.success) {
                    ctx.addIssue({
                      code: "custom",
                      message: z.treeifyError(spHostId.error).errors.join(),
                      path: ["hostId"],
                    });
                  }
                }
                break;
            }
          }
          break;

        case DEPLOY_TARGET_CERTIFICATE:
          {
            const scCertificateId = z.coerce.number().int().positive();
            const spCertificateId = scCertificateId.safeParse(values.certificateId);
            if (!spCertificateId.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spCertificateId.error).errors.join(),
                path: ["certificateId"],
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
