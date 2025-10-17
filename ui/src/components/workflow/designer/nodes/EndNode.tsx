import { getI18n } from "react-i18next";
import { FlowNodeBaseType } from "@flowgram.ai/fixed-layout-editor";
import { IconLogout } from "@tabler/icons-react";

import { newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";

export const EndNodeRegistry: NodeRegistry = {
  type: NodeType.End,

  kind: NodeKindType.Basis,

  meta: {
    labelText: getI18n().t("workflow_node.end.label"),

    icon: IconLogout,
    iconColor: "#fff",
    iconBgColor: "#336df4",

    isNodeEnd: true,

    clickable: false,
    expandable: false,
    selectable: false,

    copyDisable: true,
  },

  formMeta: {
    render: () => {
      return <BaseNode />;
    },
  },

  canAdd(_, from) {
    // You can only add to the last node of the branch
    if (!from.isLast) return false;

    // `originParent` can determine whether it is condition, and then determine whether it is the last one
    // https://github.com/bytedance/flowgram.ai/pull/146
    if (from.parent && from.parent.parent?.flowNodeType === FlowNodeBaseType.INLINE_BLOCKS && from.parent.originParent && !from.parent.originParent.isLast) {
      const allBranches = from.parent.parent!.blocks;
      // Determine whether the last node of all branch is end, all branches are not allowed to be end
      const branchEndCount = allBranches.filter((block) => block.blocks[block.blocks.length - 1]?.getNodeMeta().isNodeEnd).length;
      return branchEndCount < allBranches.length - 1;
    }

    return true;
  },

  canDelete(ctx, node) {
    return node.parent !== ctx.document.root;
  },

  onAdd() {
    return newNode(NodeType.End, { i18n: getI18n() });
  },
};
