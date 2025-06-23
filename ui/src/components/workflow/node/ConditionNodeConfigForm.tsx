import { forwardRef, memo, useImperativeHandle, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Form, type FormInstance } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import { type Expr, type WorkflowNodeConfigForCondition, defaultNodeConfigForCondition } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";

import ConditionNodeConfigFormExpressionEditor, { type ConditionNodeConfigFormExpressionEditorInstance } from "./ConditionNodeConfigFormExpressionEditor";

export type ConditionNodeConfigFormFieldValues = {
  expression?: Expr | undefined;
};

export type ConditionNodeConfigFormProps = {
  className?: string;
  style?: React.CSSProperties;
  disabled?: boolean;
  initialValues?: Partial<WorkflowNodeConfigForCondition>;
  nodeId: string;
  onValuesChange?: (values: WorkflowNodeConfigForCondition) => void;
};

export type ConditionNodeConfigFormInstance = {
  getFieldsValue: () => ReturnType<FormInstance<ConditionNodeConfigFormFieldValues>["getFieldsValue"]>;
  resetFields: FormInstance<ConditionNodeConfigFormFieldValues>["resetFields"];
  validateFields: FormInstance<ConditionNodeConfigFormFieldValues>["validateFields"];
};

const initFormModel = (): ConditionNodeConfigFormFieldValues => {
  return defaultNodeConfigForCondition();
};

const ConditionNodeConfigForm = forwardRef<ConditionNodeConfigFormInstance, ConditionNodeConfigFormProps>(
  ({ className, style, disabled, initialValues, nodeId, onValuesChange }, ref) => {
    const { t } = useTranslation();

    const formSchema = z.object({
      expression: z.any().nullish(),
    });
    const formRule = createSchemaFieldRule(formSchema);
    const { form: formInst, formProps } = useAntdForm({
      name: "workflowNodeConditionConfigForm",
      initialValues: initialValues ?? initFormModel(),
    });

    const editorRef = useRef<ConditionNodeConfigFormExpressionEditorInstance>(null);

    const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
      onValuesChange?.(values);
    };

    useImperativeHandle(ref, () => {
      return {
        getFieldsValue: formInst.getFieldsValue,
        resetFields: formInst.resetFields,
        validateFields: (nameList, config) => {
          const t1 = formInst.validateFields(nameList, config);
          const t2 = editorRef.current!.validate();
          return Promise.all([t1, t2]).then(() => t1);
        },
      } as ConditionNodeConfigFormInstance;
    });

    return (
      <Form className={className} style={style} {...formProps} disabled={disabled} layout="vertical" scrollToFirstError onValuesChange={handleFormChange}>
        <Form.Item name="expression" label={t("workflow_node.condition.form.expression.label")} rules={[formRule]}>
          <ConditionNodeConfigFormExpressionEditor ref={editorRef} nodeId={nodeId} />
        </Form.Item>
      </Form>
    );
  }
);

export default memo(ConditionNodeConfigForm);
