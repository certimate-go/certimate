import { getI18n } from "react-i18next";
import { FeedbackLevel, Field, FlowNodeBaseType, FlowNodeSplitType } from "@flowgram.ai/fixed-layout-editor";
import { IconFilter, IconFilterFilled, IconSitemap } from "@tabler/icons-react";
import { Typography } from "antd";

import { type Expr, ExprType, newNode } from "@/domain/workflow";

import { getAllPreviousNodes } from "../_util";
import { BaseNode, BranchNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BranchBlockNodeConfigForm from "../forms/BranchBlockNodeConfigForm";

export const ConditionNodeRegistry: NodeRegistry = {
  type: NodeType.Condition,

  kind: NodeKindType.Logic,

  extend: FlowNodeSplitType.DYNAMIC_SPLIT,

  meta: {
    labelText: getI18n().t("workflow_node.condition.label"),

    icon: IconSitemap,
    iconColor: "#fff",
    iconBgColor: "#373c43",

    clickable: false,
    expandable: false,

    deleteDisable: false,
  },

  formMeta: {
    render: () => {
      return <BaseNode />;
    },
  },

  onAdd() {
    return newNode(NodeType.Condition, { i18n: getI18n() });
  },
};

export const BranchBlockNodeRegistry: NodeRegistry = {
  type: NodeType.BranchBlock,

  kind: NodeKindType.Logic,

  extend: FlowNodeBaseType.BLOCK,

  meta: {
    labelText: getI18n().t("workflow_node.branch_block.label"),

    icon: IconSitemap,
    iconColor: "#fff",
    iconBgColor: "#373c43",

    clickable: true,

    addDisable: true,
    copyDisable: true,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BranchBlockNodeConfigForm.getSchema({}).safeParse(value);
        if (!res.success) {
          return {
            message: res.error.message,
            level: FeedbackLevel.Error,
          };
        }
      },
      ["config.expression"]: ({ value, context: { node } }) => {
        if (value == null) return;

        const prevNodeIds = getAllPreviousNodes(node).map((e) => e.id);
        const deepValidate = (expr: Expr) => {
          if ("selector" in expr) {
            if (!prevNodeIds.includes(expr.selector.id)) {
              return false;
            }
          }

          if ("left" in expr) {
            if (!deepValidate(expr.left)) {
              return false;
            }
          }

          if ("right" in expr) {
            if (!deepValidate(expr.right)) {
              return false;
            }
          }

          return true;
        };
        if (!deepValidate(value)) {
          return {
            message: "Invalid input",
            level: FeedbackLevel.Error,
          };
        }
      },
    },

    render: () => {
      const { t } = getI18n();

      return (
        <BranchNode
          description={
            <>
              <div className="flex items-center justify-center gap-2">
                <div className="flex items-center justify-center">
                  <Field<Expr> name="config.expression">
                    {({ field: { value } }) => (
                      <>
                        {value == null ? (
                          <IconFilter size="1.25em" stroke="1.25" />
                        ) : (
                          <IconFilterFilled color="var(--color-primary)" size="1.25em" stroke="1.25" />
                        )}
                      </>
                    )}
                  </Field>
                </div>
                <div className="truncate">
                  <Field<string> name="name">{({ field: { value } }) => <>{value || "\u00A0"}</>}</Field>
                </div>
              </div>
              <div className="mt-1">
                <div className="truncate">
                  <Field<Expr> name="config.expression">
                    {({ field: { value } }) => (
                      <Typography.Text className="text-xs" type="secondary">
                        {value == null
                          ? t("workflow_node.branch_block.state.no")
                          : value.type === ExprType.Logical && value.operator === "and"
                            ? t("workflow_node.branch_block.state.and")
                            : t("workflow_node.branch_block.state.or")}
                      </Typography.Text>
                    )}
                  </Field>
                </div>
              </div>
            </>
          }
        />
      );
    },
  },

  canAdd: () => {
    return false;
  },

  canDelete: (_, node) => {
    return node.parent != null && node.parent.blocks.length >= 2;
  },

  onAdd(_, from) {
    const node = newNode(NodeType.BranchBlock, { i18n: getI18n() });
    if (from != null) {
      const siblingLength = from.blocks?.find((b) => b.isInlineBlocks)?.blocks?.length;
      if (siblingLength != null) {
        node.data.name = `${node.data.name} ${siblingLength + 1}`;
      }
    }

    return node;
  },
};
