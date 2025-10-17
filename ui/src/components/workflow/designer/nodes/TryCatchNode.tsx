import { getI18n } from "react-i18next";
import { Field } from "@flowgram.ai/fixed-layout-editor";
import { IconArrowsSplit, IconCircleX } from "@tabler/icons-react";

import { newNode } from "@/domain/workflow";

import { BaseNode, BranchNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";

export const TryCatchNodeRegistry: NodeRegistry = {
  type: NodeType.TryCatch,

  kind: NodeKindType.Logic,

  meta: {
    labelText: getI18n().t("workflow_node.try_catch.label"),

    icon: IconArrowsSplit,
    iconColor: "#fff",
    iconBgColor: "#373c43",

    clickable: false,
    expandable: false,
  },

  formMeta: {
    render: () => {
      return <BaseNode />;
    },
  },

  onAdd() {
    return newNode(NodeType.TryCatch, { i18n: getI18n() });
  },
};

export const CatchBlockNodeRegistry: NodeRegistry = {
  type: NodeType.CatchBlock,

  kind: NodeKindType.Logic,

  meta: {
    labelText: getI18n().t("workflow_node.catch_block.label"),

    clickable: false,
    draggable: false,

    addDisable: true,
    copyDisable: true,
  },

  formMeta: {
    render: () => {
      return (
        <BranchNode
          description={
            <div className="flex items-center justify-center gap-2">
              <div className="flex items-center justify-center">
                <IconCircleX color="var(--color-error)" size="1.25em" stroke="1.25" />
              </div>
              <div className="truncate">
                <Field<string> name="name">{({ field: { value } }) => <>{value || "\u00A0"}</>}</Field>
              </div>
            </div>
          }
        />
      );
    },
  },

  canAdd: () => false,

  canDelete: (_, node) => {
    return node.parent != null && node.parent.blocks.length >= 2;
  },

  onAdd() {
    return newNode(NodeType.CatchBlock, { i18n: getI18n() });
  },
};
