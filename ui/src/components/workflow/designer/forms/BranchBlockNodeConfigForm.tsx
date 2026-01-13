import { useMemo, useRef } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";
import { type AnchorProps, Form, type FormInstance } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import {
  type Expr,
  type ExprComparisonOperator,
  type ExprLogicalOperator,
  ExprType,
  type ExprValueType,
  type WorkflowNodeConfigForBranchBlock,
  defaultNodeConfigForBranchBlock,
} from "@/domain/workflow";

import { useAntdForm } from "@/hooks";

import { NodeFormContextProvider } from "./_context";
import BranchBlockNodeConfigExprInputBox, { type BranchBlockNodeConfigExprInputBoxInstance } from "./BranchBlockNodeConfigExprInputBox";

import { NodeType } from "../nodes/typings";

export interface BranchBlockNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BranchBlockNodeConfigForm = ({ node, ...props }: BranchBlockNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BranchBlock) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BranchBlock}`);
  }

  const { i18n, t } = useTranslation();

  const initialValues = useMemo(() => {
    return node.form?.getValueIn("config") as WorkflowNodeConfigForBranchBlock | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n }).superRefine(async (values, ctx) => {
    if (values.expression != null) {
      try {
        await exprInputBoxRef.current!.validate();
      } catch {
        if (!ctx.issues.some((issue) => issue.path?.[0] === "expression")) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.branch_block.form.expression.errmsg.invalid"),
            path: ["expression"],
          });
        }
      }
    }
  });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBranchBlockConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const exprInputBoxRef = useRef<BranchBlockNodeConfigExprInputBoxInstance>(null);

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <div id="parameters" data-anchor="parameters">
          <Form.Item name="expression" label={t("workflow_node.branch_block.form.expression.label")} rules={[formRule]} validateTrigger={false}>
            <BranchBlockNodeConfigExprInputBox ref={exprInputBoxRef} />
          </Form.Item>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters"].map((key) => ({
    key: key,
    title: t(`workflow_node.branch_block.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return defaultNodeConfigForBranchBlock();
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  const exprSchema: z.ZodType<Expr> = z.lazy(() =>
    z.discriminatedUnion("type", [
      z.object({
        type: z.literal(ExprType.Constant),
        value: z.string(),
        valueType: z.string<ExprValueType>(),
      }),

      z.object({
        type: z.literal(ExprType.Variant),
        selector: z.object({
          id: z.string(),
          name: z.string(),
          type: z.string<ExprValueType>(),
        }),
      }),

      z.object({
        type: z.literal(ExprType.Comparison),
        operator: z.string<ExprComparisonOperator>(),
        left: exprSchema,
        right: exprSchema,
      }),

      z.object({
        type: z.literal(ExprType.Logical),
        operator: z.string<ExprLogicalOperator>(),
        left: exprSchema,
        right: exprSchema,
      }),

      z.object({
        type: z.literal(ExprType.Not),
        expr: exprSchema,
      }),
    ])
  );

  return z.object({
    expression: z
      .any()
      .nullish()
      .refine((v) => v == null || exprSchema.safeParse(v).success, t("workflow_node.branch_block.form.expression.errmsg.invalid")),
  });
};

const _default = Object.assign(BranchBlockNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
