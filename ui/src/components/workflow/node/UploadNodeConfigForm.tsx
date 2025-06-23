import { forwardRef, memo, useImperativeHandle } from "react";
import { useTranslation } from "react-i18next";
import { Form, type FormInstance, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { validateCertificate, validatePrivateKey } from "@/api/certificates";
import TextFileInput from "@/components/TextFileInput";
import { type WorkflowNodeConfigForUpload, defaultNodeConfigForUpload } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { getErrMsg } from "@/utils/error";

type UploadNodeConfigFormFieldValues = Partial<WorkflowNodeConfigForUpload>;

export type UploadNodeConfigFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: UploadNodeConfigFormFieldValues;
  onValuesChange?: (values: UploadNodeConfigFormFieldValues) => void;
};

export type UploadNodeConfigFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<UploadNodeConfigFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<UploadNodeConfigFormFieldValues>["resetFields"];
  validateFields: FormInstance<UploadNodeConfigFormFieldValues>["validateFields"];
};

const initFormModel = (): UploadNodeConfigFormFieldValues => {
  return defaultNodeConfigForUpload();
};

const UploadNodeConfigForm = forwardRef<UploadNodeConfigFormInstance, UploadNodeConfigFormProps>(
  ({ className, style, disabled, initialValues, onValuesChange }, ref) => {
    const { t } = useTranslation();

    const formSchema = z.object({
      domains: z.string().nullish(),
      certificate: z
        .string(t("workflow_node.upload.form.certificate.placeholder"))
        .min(1, t("workflow_node.upload.form.certificate.placeholder"))
        .max(20480, t("common.errmsg.string_max", { max: 20480 })),
      privateKey: z
        .string(t("workflow_node.upload.form.private_key.placeholder"))
        .min(1, t("workflow_node.upload.form.private_key.placeholder"))
        .max(20480, t("common.errmsg.string_max", { max: 20480 })),
    });
    const formRule = createSchemaFieldRule(formSchema);
    const { form: formInst, formProps } = useAntdForm({
      name: "workflowNodeUploadConfigForm",
      initialValues: initialValues ?? initFormModel(),
    });

    const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
      onValuesChange?.(values as UploadNodeConfigFormFieldValues);
    };

    useImperativeHandle(ref, () => {
      return {
        getFieldsValue: () => {
          return formInst.getFieldsValue(true);
        },
        resetFields: (fields) => {
          return formInst.resetFields(fields as (keyof UploadNodeConfigFormFieldValues)[]);
        },
        validateFields: (nameList, config) => {
          return formInst.validateFields(nameList, config);
        },
      } as UploadNodeConfigFormInstance;
    });

    const handleCertificateChange = async (value: string) => {
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

      onValuesChange?.(formInst.getFieldsValue(true));
    };

    const handlePrivateKeyChange = async (value: string) => {
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

      onValuesChange?.(formInst.getFieldsValue(true));
    };

    return (
      <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
        <Form.Item name="domains" label={t("workflow_node.upload.form.domains.label")} rules={[formRule]}>
          <Input variant="filled" placeholder={t("workflow_node.upload.form.domains.placeholder")} readOnly />
        </Form.Item>

        <Form.Item name="certificate" label={t("workflow_node.upload.form.certificate.label")} rules={[formRule]}>
          <TextFileInput
            autoSize={{ minRows: 3, maxRows: 10 }}
            placeholder={t("workflow_node.upload.form.certificate.placeholder")}
            onChange={handleCertificateChange}
          />
        </Form.Item>

        <Form.Item name="privateKey" label={t("workflow_node.upload.form.private_key.label")} rules={[formRule]}>
          <TextFileInput
            autoSize={{ minRows: 3, maxRows: 10 }}
            placeholder={t("workflow_node.upload.form.private_key.placeholder")}
            onChange={handlePrivateKeyChange}
          />
        </Form.Item>
      </Form>
    );
  }
);

export default memo(UploadNodeConfigForm);
