import { forwardRef, memo, useImperativeHandle } from "react";
import { useTranslation } from "react-i18next";
import { Alert, Form, type FormInstance, Input, InputNumber } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type WorkflowNodeConfigForMonitor, defaultNodeConfigForMonitor } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { validDomainName, validIPv4Address, validIPv6Address, validPortNumber } from "@/utils/validators";

type MonitorNodeConfigFormFieldValues = Partial<WorkflowNodeConfigForMonitor>;

export type MonitorNodeConfigFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: MonitorNodeConfigFormFieldValues;
  onValuesChange?: (values: MonitorNodeConfigFormFieldValues) => void;
};

export type MonitorNodeConfigFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<MonitorNodeConfigFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<MonitorNodeConfigFormFieldValues>["resetFields"];
  validateFields: FormInstance<MonitorNodeConfigFormFieldValues>["validateFields"];
};

const initFormModel = (): MonitorNodeConfigFormFieldValues => {
  return defaultNodeConfigForMonitor();
};

const MonitorNodeConfigForm = forwardRef<MonitorNodeConfigFormInstance, MonitorNodeConfigFormProps>(
  ({ className, style, disabled, initialValues, onValuesChange }, ref) => {
    const { t } = useTranslation();

    const formSchema = z.object({
      host: z.string().refine((v) => {
        return validDomainName(v) || validIPv4Address(v) || validIPv6Address(v);
      }, t("common.errmsg.host_invalid")),
      port: z.preprocess(
        (v) => Number(v),
        z
          .number()
          .int(t("workflow_node.monitor.form.port.placeholder"))
          .refine((v) => validPortNumber(v), t("common.errmsg.port_invalid"))
      ),
      domain: z
        .string()
        .nullish()
        .refine((v) => {
          if (!v) return true;
          return validDomainName(v);
        }, t("common.errmsg.domain_invalid")),
      requestPath: z.string().nullish(),
    });
    const formRule = createSchemaFieldRule(formSchema);
    const { form: formInst, formProps } = useAntdForm({
      name: "workflowNodeMonitorConfigForm",
      initialValues: initialValues ?? initFormModel(),
    });

    const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
      onValuesChange?.(values as MonitorNodeConfigFormFieldValues);
    };

    useImperativeHandle(ref, () => {
      return {
        getFieldsValue: () => {
          return formInst.getFieldsValue(true);
        },
        resetFields: (fields) => {
          return formInst.resetFields(fields as (keyof MonitorNodeConfigFormFieldValues)[]);
        },
        validateFields: (nameList, config) => {
          return formInst.validateFields(nameList, config);
        },
      } as MonitorNodeConfigFormInstance;
    });

    return (
      <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
        <Form.Item>
          <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.monitor.form.guide") }}></span>} />
        </Form.Item>

        <div className="flex space-x-2">
          <div className="w-2/3">
            <Form.Item name="host" label={t("workflow_node.monitor.form.host.label")} rules={[formRule]}>
              <Input placeholder={t("workflow_node.monitor.form.host.placeholder")} />
            </Form.Item>
          </div>

          <div className="w-1/3">
            <Form.Item name="port" label={t("workflow_node.monitor.form.port.label")} rules={[formRule]}>
              <InputNumber style={{ width: "100%" }} min={1} max={65535} placeholder={t("workflow_node.monitor.form.port.placeholder")} />
            </Form.Item>
          </div>
        </div>

        <Form.Item name="domain" label={t("workflow_node.monitor.form.domain.label")} rules={[formRule]}>
          <Input placeholder={t("workflow_node.monitor.form.domain.placeholder")} />
        </Form.Item>

        <Form.Item name="requestPath" label={t("workflow_node.monitor.form.request_path.label")} rules={[formRule]}>
          <Input placeholder={t("workflow_node.monitor.form.request_path.placeholder")} />
        </Form.Item>
      </Form>
    );
  }
);

export default memo(MonitorNodeConfigForm);
