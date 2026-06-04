import { getI18n, useTranslation } from "react-i18next";
import { IconDice6 } from "@tabler/icons-react";
import { Button, Form, Input, Select, Space, Tooltip } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";

import { useFormNestedFieldsContext } from "./_context";
import { useSharedFormFieldsAndHandlers as _useSharedFormFieldsAndHandlers } from "./BizDeployNodeConfigFieldsProviderLocal";

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

  const {
    fieldFormat: fieldFileFormat,
    handleChangeFormat: handleFileFormatSelect,
    handleRandomPfxPassword: handleRandomPfxPasswordClick,
    handleRandomJksAlias: handleRandomJksAliasClick,
    handleRandomJksKeypass: handleRandomJksKeypassClick,
    handleRandomJksStorepass: handleRandomJksStorepassClick,
  } = _useSharedFormFieldsAndHandlers(formInst, { certPath: "objectKeyForCrt" });

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
        name={[parentNamePath, "fileFormat"]}
        initialValue={initialValues.fileFormat}
        label={t("workflow_node.deploy.form.shared_file_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[FORMAT_PEM, FORMAT_PFX, FORMAT_JKS].map((s) => ({
            label: t(`workflow_node.deploy.form.shared_file_format.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_file_format.placeholder")}
          onSelect={handleFileFormatSelect}
        />
      </Form.Item>

      <Show when={fieldFileFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "objectKeyForKey"]}
          initialValue={initialValues.objectKeyForKey}
          label={t("workflow_node.deploy.form.s3_object_key_for_key.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.s3_object_key_for_key.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "objectKeyForCrt"]}
        initialValue={initialValues.objectKeyForCrt}
        label={t(`workflow_node.deploy.form.s3_object_key_for_crt.label`)}
        rules={[formRule]}
      >
        <Input placeholder={t(`workflow_node.deploy.form.s3_object_key_for_crt.placeholder`)} />
      </Form.Item>

      <Show when={fieldFileFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "objectKeyForCrtOnlyServer"]}
          initialValue={initialValues.objectKeyForCrtOnlyServer}
          label={t("workflow_node.deploy.form.s3_object_key_for_servercrt.label")}
          extra={t("workflow_node.deploy.form.s3_object_key_for_servercrt.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.s3_object_key_for_servercrt.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "objectKeyForCrtOnlyIntermedia"]}
          initialValue={initialValues.objectKeyForCrtOnlyIntermedia}
          label={t("workflow_node.deploy.form.s3_object_key_for_intermediacrt.label")}
          extra={t("workflow_node.deploy.form.s3_object_key_for_intermediacrt.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.s3_object_key_for_intermediacrt.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFileFormat === FORMAT_PFX}>
        <Form.Item label={t("workflow_node.deploy.form.shared_pfx_password.label")}>
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "pfxPassword"]} initialValue={initialValues.pfxPassword} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_pfx_password.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomPfxPasswordClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "pfxEncoder"]}
          initialValue={initialValues.pfxEncoder}
          label={t("workflow_node.deploy.form.shared_pfx_encoder.label")}
          rules={[formRule]}
        >
          <Select
            options={["LegacyRC2", "LegacyDES", "Modern2023", "Modern2026"].map((s) => ({
              label: t(`workflow_node.deploy.form.shared_pfx_encoder.option.${s.toLowerCase()}.label`),
              value: s,
            }))}
            placeholder={t("workflow_node.deploy.form.shared_pfx_encoder.placeholder")}
          />
        </Form.Item>
      </Show>

      <Show when={fieldFileFormat === FORMAT_JKS}>
        <Form.Item
          label={t("workflow_node.deploy.form.shared_jks_alias.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_jks_alias.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksAlias"]} initialValue={initialValues.jksAlias} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_jks_alias.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksAliasClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          label={t("workflow_node.deploy.form.shared_jks_keypass.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_jks_keypass.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksKeypass"]} initialValue={initialValues.jksKeypass} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_jks_keypass.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksKeypassClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          label={t("workflow_node.deploy.form.shared_jks_storepass.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_jks_storepass.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksStorepass"]} initialValue={initialValues.jksStorepass} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_jks_storepass.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksStorepassClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    region: "",
    bucket: "",
    fileFormat: FORMAT_PEM,
    objectKeyForKey: ".certimate/cert.key",
    objectKeyForCrt: ".certimate/cert.crt",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      region: z.string().nonempty(),
      bucket: z.string().nonempty(),
      fileFormat: z.enum([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS]),
      objectKeyForKey: z.string().max(256).nullish(),
      objectKeyForCrt: z.string().max(256).nullish(),
      objectKeyForCrtOnlyServer: z.string().max(256).nullish(),
      objectKeyForCrtOnlyIntermedia: z.string().max(256).nullish(),
      pfxPassword: z.string().nullish(),
      pfxEncoder: z.string().nullish(),
      jksAlias: z.string().nullish(),
      jksKeypass: z.string().nullish(),
      jksStorepass: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.fileFormat) {
        case FORMAT_PFX:
          {
            const scPfxPassword = z.string().nonempty();
            const spPfxPassword = scPfxPassword.safeParse(values.pfxPassword);
            if (!spPfxPassword.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPfxPassword.error).errors.join(),
                path: ["pfxPassword"],
              });
            }
          }
          break;

        case FORMAT_JKS:
          {
            const scJksAlias = z.string().nonempty();
            const spJksAlias = scJksAlias.safeParse(values.jksAlias);
            if (!spJksAlias.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spJksAlias.error).errors.join(),
                path: ["jksAlias"],
              });
            }

            const scJksKeypass = z.string().nonempty();
            const spJksKeypass = scJksKeypass.safeParse(values.jksKeypass);
            if (!spJksKeypass.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spJksKeypass.error).errors.join(),
                path: ["jksKeypass"],
              });
            }

            const scJksStorepass = z.string().nonempty();
            const spJksStorepass = scJksStorepass.safeParse(values.jksStorepass);
            if (!spJksStorepass.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spJksStorepass.error).errors.join(),
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
