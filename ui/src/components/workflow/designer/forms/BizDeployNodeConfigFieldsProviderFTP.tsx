import { getI18n, useTranslation } from "react-i18next";
import { IconDice6 } from "@tabler/icons-react";
import { Button, Form, Input, Select, Space, Tooltip } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";
import { randomString } from "@/utils/random";

import { useFormNestedFieldsContext } from "./_context";
import { initPresetScript as _initPresetScript } from "./BizDeployNodeConfigFieldsProviderLocal";

const FORMAT_PEM = CERTIFICATE_FORMATS.PEM;
const FORMAT_PFX = CERTIFICATE_FORMATS.PFX;
const FORMAT_JKS = CERTIFICATE_FORMATS.JKS;

const BizDeployNodeConfigFieldsProviderFTP = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldFormat = Form.useWatch([parentNamePath, "format"], formInst);
  const fieldCertPath = Form.useWatch([parentNamePath, "certPath"], formInst);

  const handleFormatSelect = (value: string) => {
    if (fieldFormat === value) return;

    switch (value) {
      case FORMAT_PEM:
        {
          if (/(.pfx|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certPath"], fieldCertPath.replace(/(.pfx|.jks)$/, ".crt"));
          }
        }
        break;

      case FORMAT_PFX:
        {
          if (/(.crt|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certPath"], fieldCertPath.replace(/(.crt|.jks)$/, ".pfx"));
          }
        }
        break;

      case FORMAT_JKS:
        {
          if (/(.crt|.pfx)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certPath"], fieldCertPath.replace(/(.crt|.pfx)$/, ".jks"));
          }
        }
        break;
    }
  };

  const handleRandomPfxPasswordClick = () => {
    const password = randomString();
    formInst.setFieldValue([parentNamePath, "pfxPassword"], password);
  };

  const handleRandomJksKeypassClick = () => {
    const password = randomString();
    formInst.setFieldValue([parentNamePath, "jksKeypass"], password);
  };

  const handleRandomJksStorepassClick = () => {
    const password = randomString();
    formInst.setFieldValue([parentNamePath, "jksStorepass"], password);
  };

  return (
    <>
      <Form.Item
        name={[parentNamePath, "format"]}
        initialValue={initialValues.format}
        label={t("workflow_node.deploy.form.ftp_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[FORMAT_PEM, FORMAT_PFX, FORMAT_JKS].map((s) => ({
            key: s,
            label: t(`workflow_node.deploy.form.ftp_format.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.ftp_format.placeholder")}
          onSelect={handleFormatSelect}
        />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "keyPath"]}
          initialValue={initialValues.keyPath}
          label={t("workflow_node.deploy.form.ftp_key_path.label")}
          extra={t("workflow_node.deploy.form.ftp_key_path.help")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.ftp_key_path.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "certPath"]}
        initialValue={initialValues.certPath}
        label={t(`workflow_node.deploy.form.ftp_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_path.label`)}
        extra={t("workflow_node.deploy.form.ftp_cert_path.help")}
        rules={[formRule]}
      >
        <Input placeholder={t(`workflow_node.deploy.form.ftp_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_path.placeholder`)} />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "certPathForServerOnly"]}
          initialValue={initialValues.certPathForServerOnly}
          label={t("workflow_node.deploy.form.ftp_servercert_path.label")}
          extra={t("workflow_node.deploy.form.ftp_servercert_path.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.ftp_servercert_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "certPathForIntermediaOnly"]}
          initialValue={initialValues.certPathForIntermediaOnly}
          label={t("workflow_node.deploy.form.ftp_intermediacert_path.label")}
          extra={t("workflow_node.deploy.form.ftp_intermediacert_path.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.ftp_intermediacert_path.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_PFX}>
        <Form.Item
          label={t("workflow_node.deploy.form.ftp_pfx_password.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ftp_pfx_password.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "pfxPassword"]} initialValue={initialValues.pfxPassword} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.ftp_pfx_password.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomPfxPasswordClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_JKS}>
        <Form.Item
          name={[parentNamePath, "jksAlias"]}
          initialValue={initialValues.jksAlias}
          label={t("workflow_node.deploy.form.ftp_jks_alias.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ftp_jks_alias.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ftp_jks_alias.placeholder")} />
        </Form.Item>

        <Form.Item
          label={t("workflow_node.deploy.form.ftp_jks_keypass.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ftp_jks_keypass.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksKeypass"]} initialValue={initialValues.jksKeypass} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.ftp_jks_keypass.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksKeypassClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          label={t("workflow_node.deploy.form.ftp_jks_storepass.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ftp_jks_storepass.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksStorepass"]} initialValue={initialValues.jksStorepass} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.ftp_jks_storepass.placeholder")} />
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
    format: FORMAT_PEM,
    keyPath: "/certimate/cert.key",
    certPath: "/certimate/cert.crt",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      format: z.literal([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS], t("workflow_node.deploy.form.ftp_format.placeholder")),
      keyPath: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certPath: z
        .string()
        .min(1, t("workflow_node.deploy.form.ftp_cert_path.placeholder"))
        .max(256, t("common.errmsg.string_max", { max: 256 })),
      certPathForServerOnly: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certPathForIntermediaOnly: z
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
            if (!values.keyPath?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ftp_key_path.placeholder"),
                path: ["keyPath"],
              });
            }
          }
          break;

        case FORMAT_PFX:
          {
            if (!values.pfxPassword?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ftp_pfx_password.placeholder"),
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
                message: t("workflow_node.deploy.form.ftp_jks_alias.placeholder"),
                path: ["jksAlias"],
              });
            }

            if (!values.jksKeypass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ftp_jks_keypass.placeholder"),
                path: ["jksKeypass"],
              });
            }

            if (!values.jksStorepass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ftp_jks_storepass.placeholder"),
                path: ["jksStorepass"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderFTP, {
  getInitialValues,
  getSchema,
});

export default _default;
