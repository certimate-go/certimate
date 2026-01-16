import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";

import { useFormNestedFieldsContext } from "./_context";

const FORMAT_PEM = CERTIFICATE_FORMATS.PEM;
const FORMAT_PFX = CERTIFICATE_FORMATS.PFX;
const FORMAT_JKS = CERTIFICATE_FORMATS.JKS;

const BizDeployNodeConfigFieldsProviderS3 = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldFormat = Form.useWatch([parentNamePath, "format"], formInst);
  const fieldCertPath = Form.useWatch([parentNamePath, "certObjectKey"], formInst);

  const handleFormatSelect = (value: string) => {
    if (fieldFormat === value) return;

    switch (value) {
      case FORMAT_PEM:
        {
          if (/(.pfx|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certObjectKey"], fieldCertPath.replace(/(.pfx|.jks)$/, ".crt"));
          }
        }
        break;

      case FORMAT_PFX:
        {
          if (/(.crt|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certObjectKey"], fieldCertPath.replace(/(.crt|.jks)$/, ".pfx"));
          }
        }
        break;

      case FORMAT_JKS:
        {
          if (/(.crt|.pfx)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certObjectKey"], fieldCertPath.replace(/(.crt|.pfx)$/, ".jks"));
          }
        }
        break;
    }
  };

  return (
    <>
      <Form.Item
        name={[parentNamePath, "region"]}
        initialValue={initialValues.region}
        label={t("workflow_node.deploy.form.s3_region.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.s3_region.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "bucket"]}
        initialValue={initialValues.bucket}
        label={t("workflow_node.deploy.form.s3_bucket.label")}
        rules={[formRule]}
      >
        <Input placeholder={t("workflow_node.deploy.form.s3_bucket.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "format"]}
        initialValue={initialValues.format}
        label={t("workflow_node.deploy.form.s3_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[FORMAT_PEM, FORMAT_PFX, FORMAT_JKS].map((s) => ({
            key: s,
            label: t(`workflow_node.deploy.form.s3_format.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.s3_format.placeholder")}
          onSelect={handleFormatSelect}
        />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "keyObjectKey"]}
          initialValue={initialValues.keyObjectKey}
          label={t("workflow_node.deploy.form.s3_key_object_key.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.s3_key_object_key.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "certObjectKey"]}
        initialValue={initialValues.certObjectKey}
        label={t(`workflow_node.deploy.form.s3_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_object_key.label`)}
        rules={[formRule]}
      >
        <Input placeholder={t(`workflow_node.deploy.form.s3_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_object_key.placeholder`)} />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "certObjectKeyForServerOnly"]}
          initialValue={initialValues.certObjectKeyForServerOnly}
          label={t("workflow_node.deploy.form.s3_servercert_object_key.label")}
          extra={t("workflow_node.deploy.form.s3_servercert_object_key.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.s3_servercert_object_key.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "certObjectKeyForIntermediaOnly"]}
          initialValue={initialValues.certObjectKeyForIntermediaOnly}
          label={t("workflow_node.deploy.form.s3_intermediacert_object_key.label")}
          extra={t("workflow_node.deploy.form.s3_intermediacert_object_key.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.s3_intermediacert_object_key.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_PFX}>
        <Form.Item
          name={[parentNamePath, "pfxPassword"]}
          initialValue={initialValues.pfxPassword}
          label={t("workflow_node.deploy.form.s3_pfx_password.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.s3_pfx_password.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.s3_pfx_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_JKS}>
        <Form.Item
          name={[parentNamePath, "jksAlias"]}
          initialValue={initialValues.jksAlias}
          label={t("workflow_node.deploy.form.s3_jks_alias.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.s3_jks_alias.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.s3_jks_alias.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "jksKeypass"]}
          initialValue={initialValues.jksKeypass}
          label={t("workflow_node.deploy.form.s3_jks_keypass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.s3_jks_keypass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.s3_jks_keypass.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "jksStorepass"]}
          initialValue={initialValues.jksStorepass}
          label={t("workflow_node.deploy.form.s3_jks_storepass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.s3_jks_storepass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.s3_jks_storepass.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    bucket: "",
    format: FORMAT_PEM,
    keyObjectKey: ".certimate/cert.key",
    certObjectKey: ".certimate/cert.crt",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      region: z.string().nonempty(t("workflow_node.deploy.form.s3_region.placeholder")),
      bucket: z.string().nonempty(t("workflow_node.deploy.form.s3_bucket.placeholder")),
      format: z.literal([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS], t("workflow_node.deploy.form.s3_format.placeholder")),
      keyObjectKey: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certObjectKey: z
        .string()
        .min(1, t("workflow_node.deploy.form.s3_cert_object_key.placeholder"))
        .max(256, t("common.errmsg.string_max", { max: 256 })),
      certObjectKeyForServerOnly: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certObjectKeyForIntermediaOnly: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      pfxPassword: z.string().nullish(),
      jksAlias: z.string().nullish(),
      jksKeypass: z.string().nullish(),
      jksStorepass: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.format) {
        case FORMAT_PEM:
          {
            if (!values.keyObjectKey?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.s3_key_object_key.placeholder"),
                path: ["keyObjectKey"],
              });
            }
          }
          break;

        case FORMAT_PFX:
          {
            if (!values.pfxPassword?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.s3_pfx_password.placeholder"),
                path: ["pfxPassword"],
              });
            }
          }
          break;

        case FORMAT_JKS:
          {
            if (!values.jksAlias?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.s3_jks_alias.placeholder"),
                path: ["jksAlias"],
              });
            }

            if (!values.jksKeypass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.s3_jks_keypass.placeholder"),
                path: ["jksKeypass"],
              });
            }

            if (!values.jksStorepass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.s3_jks_storepass.placeholder"),
                path: ["jksStorepass"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderS3, {
  getInitialValues,
  getSchema,
});

export default _default;
