import { useMemo } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { type AnchorProps, Form, type FormInstance, Input, Radio } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import { validateCertificate, validatePrivateKey } from "@/api/certificates";
import Show from "@/components/Show";
import FileTextInput from "@/components/FileTextInput";
import Tips from "@/components/Tips";
import { type WorkflowNodeConfigForBizUpload, defaultNodeConfigForBizUpload } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { getErrMsg } from "@/utils/error";

import { NodeFormContextProvider } from "./_context";
import { NodeType } from "../nodes/typings";

export interface BizUploadNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const UPLOAD_SOURCE_FORM = "form" as const;
const UPLOAD_SOURCE_LOCAL = "local" as const;
const UPLOAD_SOURCE_URL = "url" as const;

const BizUploadNodeConfigForm = ({ node, ...props }: BizUploadNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizUpload) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizUpload}`);
  }

  const { i18n, t } = useTranslation();

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBizUpload | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizUploadConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldSource = Form.useWatch("source", { form: formInst, preserve: true });

  const handleSourceChange = (value: string) => {
    if (value === initialValues?.source) {
      formInst.resetFields(["certificate", "privateKey", "domains"]);
    } else {
      setTimeout(() => {
        formInst.setFieldValue("certificate", "");
        formInst.setFieldValue("privateKey", "");
        formInst.setFieldValue("domains", "");
      }, 0);
    }
  };

  const handleCertificatePEMChange = async (value: string) => {
    try {
      const resp = await validateCertificate(value);
      formInst.setFields([
        {
          name: "domains",
          value: resp.data.domains,
        },
        {
          name: "certificate",
          value: value,
        },
      ]);
    } catch (e) {
      formInst.setFields([
        {
          name: "domains",
          value: "",
        },
        {
          name: "certificate",
          value: value,
          errors: [getErrMsg(e)],
        },
      ]);
    }
  };

  const handlePrivateKeyPEMChange = async (value: string) => {
    try {
      await validatePrivateKey(value);
      formInst.setFields([
        {
          name: "privateKey",
          value: value,
        },
      ]);
    } catch (e) {
      formInst.setFields([
        {
          name: "privateKey",
          value: value,
          errors: [getErrMsg(e)],
        },
      ]);
    }
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item name="source" label={t("workflow_node.upload.form.source.label")} rules={[formRule]}>
            <Radio.Group block onChange={(e) => handleSourceChange(e.target.value)}>
              <Radio.Button value={UPLOAD_SOURCE_FORM}>{t("workflow_node.upload.form.source.option.form.label")}</Radio.Button>
              <Radio.Button value={UPLOAD_SOURCE_LOCAL}>{t("workflow_node.upload.form.source.option.local.label")}</Radio.Button>
              <Radio.Button value={UPLOAD_SOURCE_URL}>{t("workflow_node.upload.form.source.option.url.label")}</Radio.Button>
            </Radio.Group>
          </Form.Item>

          <Show when={fieldSource === UPLOAD_SOURCE_FORM}>
            <Form.Item name="domains" label={t("workflow_node.upload.form.domains.label")} rules={[formRule]}>
              <Input variant="filled" placeholder={t("workflow_node.upload.form.domains.placeholder")} readOnly />
            </Form.Item>

            <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate_pem.label")} rules={[formRule]}>
              <FileTextInput
                autoSize={{ minRows: 3, maxRows: 10 }}
                placeholder={t("workflow_node.upload.form.certificate_pem.placeholder")}
                onChange={handleCertificatePEMChange}
              />
            </Form.Item>

            <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key_pem.label")} rules={[formRule]}>
              <FileTextInput
                autoSize={{ minRows: 3, maxRows: 10 }}
                placeholder={t("workflow_node.upload.form.private_key_pem.placeholder")}
                onChange={handlePrivateKeyPEMChange}
              />
            </Form.Item>
          </Show>

          <Show when={fieldSource === UPLOAD_SOURCE_LOCAL}>
            <Form.Item>
              <Tips message={t("workflow_node.upload.form.guide")} />
            </Form.Item>

            <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate_path.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.certificate_path.placeholder")} />
            </Form.Item>

            <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key_path.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.private_key_path.placeholder")} />
            </Form.Item>
          </Show>

          <Show when={fieldSource === UPLOAD_SOURCE_URL}>
            <Form.Item>
              <Tips message={t("workflow_node.upload.form.guide")} />
            </Form.Item>

            <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate_url.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.certificate_url.placeholder")} />
            </Form.Item>

            <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key_url.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.upload.form.private_key_url.placeholder")} />
            </Form.Item>
          </Show>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters"].map((key) => ({
    key: key,
    title: t(`workflow_node.upload.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    certificate: "",
    privateKey: "",
    ...(defaultNodeConfigForBizUpload() as Nullish<z.infer<ReturnType<typeof getSchema>>>),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      source: z.enum([UPLOAD_SOURCE_FORM, UPLOAD_SOURCE_LOCAL, UPLOAD_SOURCE_URL], t("workflow_node.upload.form.source.placeholder")),
      certificate: z.string().nonempty(),
      privateKey: z.string().nonempty(),
      domains: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.source) {
        case UPLOAD_SOURCE_FORM:
          {
            if (!z.string().nonempty().safeParse(values.certificate).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.upload.form.certificate_pem.placeholder"),
                path: ["certificate"],
              });
            }

            if (!z.string().nonempty().safeParse(values.privateKey).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.upload.form.private_key_pem.placeholder"),
                path: ["privateKey"],
              });
            }
          }
          break;

        case UPLOAD_SOURCE_LOCAL:
          {
            if (!z.string().nonempty().safeParse(values.certificate).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.upload.form.certificate_path.placeholder"),
                path: ["certificate"],
              });
            }

            if (!z.string().nonempty().safeParse(values.privateKey).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.upload.form.certificate_path.placeholder"),
                path: ["privateKey"],
              });
            }
          }
          break;

        case UPLOAD_SOURCE_URL:
          {
            if (!z.url().safeParse(values.certificate).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.upload.form.certificate_url.placeholder"),
                path: ["certificate"],
              });
            }

            if (!z.url().safeParse(values.privateKey).success) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.upload.form.private_key_url.placeholder"),
                path: ["privateKey"],
              });
            }
          }
          break;

        default:
          {
            ctx.addIssue({
              code: "custom",
              message: t("workflow_node.upload.form.source.placeholder"),
              path: ["source"],
            });
          }
          break;
      }
    });
};

const _default = Object.assign(BizUploadNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
