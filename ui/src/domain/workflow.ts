import { getI18n } from "react-i18next";
import { Immer } from "immer";
import { nanoid } from "nanoid";

import { type WorkflowRunModel } from "./workflowRun";

export interface WorkflowModel extends BaseModel {
  name: string;
  description?: string;
  trigger: string;
  triggerCron?: string;
  enabled?: boolean;
  graphDraft?: WorkflowGraph;
  graphContent?: WorkflowGraph;
  hasDraft?: boolean;
  hasContent?: boolean;
  lastRunRef?: string;
  lastRunStatus?: string;
  lastRunTime?: string;
  expand?: {
    lastRunRef?: Pick<WorkflowRunModel, "id" | "status" | "trigger" | "startedAt" | "endedAt" | "error">;
  };
}

export interface WorkflowGraph {
  nodes: WorkflowNode[];
}

export const WORKFLOW_TRIGGERS = Object.freeze({
  SCHEDULED: "scheduled",
  MANUAL: "manual",
} as const);

export type WorkflowTriggerType = (typeof WORKFLOW_TRIGGERS)[keyof typeof WORKFLOW_TRIGGERS];

// #region Node
export const WORKFLOW_NODE_TYPES = Object.freeze({
  START: "start",
  END: "end",
  DELAY: "delay",
  CONDITION: "condition",
  BRANCHBLOCK: "branchBlock",
  TRYCATCH: "tryCatch",
  TRYBLOCK: "tryBlock",
  CATCHBLOCK: "catchBlock",
  BIZ_APPLY: "bizApply",
  BIZ_UPLOAD: "bizUpload",
  BIZ_MONITOR: "bizMonitor",
  BIZ_DEPLOY: "bizDeploy",
  BIZ_NOTIFY: "bizNotify",
} as const);

export type WorkflowNodeType = (typeof WORKFLOW_NODE_TYPES)[keyof typeof WORKFLOW_NODE_TYPES];

export type WorkflowNode = {
  id: string;
  type: WorkflowNodeType;
  data: {
    name?: string;
    disabled?: boolean;
    config?: Record<string, unknown>;
    [key: string]: unknown;
  };
  blocks?: WorkflowNode[];
};

export type WorkflowNodeConfigForStart = {
  trigger: string;
  triggerCron?: string;
};

export const defaultNodeConfigForStart = (): Partial<WorkflowNodeConfigForStart> => {
  return {
    trigger: WORKFLOW_TRIGGERS.MANUAL,
  };
};

export type WorkflowNodeConfigForDelay = {
  wait?: number;
};

export const defaultNodeConfigForDelay = (): Partial<WorkflowNodeConfigForDelay> => {
  return {};
};

export type WorkflowNodeConfigForBranchBlock = {
  expression?: Expr;
};

export const defaultNodeConfigForBranchBlock = (): Partial<WorkflowNodeConfigForBranchBlock> => {
  return {};
};

export type WorkflowNodeConfigForBizApply = {
  identifier: "domain" | "ip";
  domains: string;
  ipaddrs: string;
  contactEmail: string;
  challengeType: string;
  provider: string;
  providerAccessId: string;
  providerConfig?: Record<string, unknown>;
  caProvider?: string;
  caProviderAccessId?: string;
  caProviderConfig?: Record<string, unknown>;
  keySource: string;
  keyAlgorithm: string;
  keyContent?: string;
  validityLifetime?: string;
  acmeProfile?: string;
  nameservers?: string;
  dnsPropagationWait?: number;
  dnsPropagationTimeout?: number;
  dnsTTL?: number;
  httpDelayWait?: number;
  disableFollowCNAME?: boolean;
  disableARI?: boolean;
  skipBeforeExpiryDays: number;
};

export const defaultNodeConfigForBizApply = (): Partial<WorkflowNodeConfigForBizApply> => {
  return {
    challengeType: "dns-01" as const,
    keySource: "auto" as const,
    keyAlgorithm: "RSA2048" as const,
    skipBeforeExpiryDays: 30,
  };
};

export type WorkflowNodeConfigForBizUpload = {
  source: string;
  certificate: string;
  privateKey: string;
};

export const defaultNodeConfigForBizUpload = (): Partial<WorkflowNodeConfigForBizUpload> => {
  return {
    source: "form" as const,
  };
};

export type WorkflowNodeConfigForBizMonitor = {
  host: string;
  port: number;
  domain?: string;
  requestPath?: string;
};

export const defaultNodeConfigForBizMonitor = (): Partial<WorkflowNodeConfigForBizMonitor> => {
  return {
    host: "",
    port: 443,
    requestPath: "/",
  };
};

export type WorkflowNodeConfigForBizDeploy = {
  certificateOutputNodeId: string;
  provider: string;
  providerAccessId?: string;
  providerConfig?: Record<string, unknown>;
  skipOnLastSucceeded: boolean;
};

export const defaultNodeConfigForBizDeploy = (): Partial<WorkflowNodeConfigForBizDeploy> => {
  return {
    skipOnLastSucceeded: true,
  };
};

export type WorkflowNodeConfigForBizNotify = {
  subject: string;
  message: string;
  provider: string;
  providerAccessId: string;
  providerConfig?: Record<string, unknown>;
  skipOnAllPrevSkipped?: boolean;
};

export const defaultNodeConfigForBizNotify = (): Partial<WorkflowNodeConfigForBizNotify> => {
  return {};
};

export const newNodeId = (): string => {
  return nanoid()
    .replace(/^[_-]+/g, "")
    .replace(/[_-]+$/g, "");
};

export const newNode = (type: WorkflowNodeType, { i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): WorkflowNode => {
  const { t } = i18n;

  switch (type) {
    case WORKFLOW_NODE_TYPES.START:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.start.default_name"),
          config: defaultNodeConfigForStart(),
        },
      };

    case WORKFLOW_NODE_TYPES.END:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.end.default_name"),
        },
      };

    case WORKFLOW_NODE_TYPES.DELAY:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.delay.default_name"),
          config: defaultNodeConfigForDelay(),
        },
      };

    case WORKFLOW_NODE_TYPES.CONDITION: {
      const branch1 = newNode(WORKFLOW_NODE_TYPES.BRANCHBLOCK, { i18n });
      branch1.data.name = `${branch1.data.name} 1`;
      const branch2 = newNode(WORKFLOW_NODE_TYPES.BRANCHBLOCK, { i18n });
      branch2.data.name = `${branch2.data.name} 2`;

      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.condition.default_name"),
          config: defaultNodeConfigForBranchBlock(),
        },
        blocks: [branch1, branch2],
      };
    }

    case WORKFLOW_NODE_TYPES.BRANCHBLOCK:
      return {
        id: newNodeId(),
        type: type,
        blocks: [],
        data: {
          name: t("workflow_node.branch_block.default_name"),
          config: defaultNodeConfigForBranchBlock(),
        },
      };

    case WORKFLOW_NODE_TYPES.TRYCATCH:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.try_catch.default_name"),
        },
        blocks: [newNode(WORKFLOW_NODE_TYPES.TRYBLOCK, { i18n }), newNode(WORKFLOW_NODE_TYPES.CATCHBLOCK, { i18n })],
      };

    case WORKFLOW_NODE_TYPES.TRYBLOCK:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: "",
        },
        blocks: [],
      };

    case WORKFLOW_NODE_TYPES.CATCHBLOCK:
      return {
        id: newNodeId(),
        type: type,
        blocks: [newNode(WORKFLOW_NODE_TYPES.END, { i18n })],
        data: {
          name: t("workflow_node.catch_block.default_name"),
        },
      };

    case WORKFLOW_NODE_TYPES.BIZ_APPLY:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.apply.default_name"),
          config: defaultNodeConfigForBizApply(),
        },
      };

    case WORKFLOW_NODE_TYPES.BIZ_UPLOAD:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.upload.default_name"),
          config: defaultNodeConfigForBizUpload(),
        },
      };

    case WORKFLOW_NODE_TYPES.BIZ_MONITOR:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.monitor.default_name"),
          config: defaultNodeConfigForBizMonitor(),
        },
      };

    case WORKFLOW_NODE_TYPES.BIZ_DEPLOY:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.deploy.default_name"),
          config: defaultNodeConfigForBizDeploy(),
        },
      };

    case WORKFLOW_NODE_TYPES.BIZ_NOTIFY:
      return {
        id: newNodeId(),
        type: type,
        data: {
          name: t("workflow_node.notify.default_name"),
          config: defaultNodeConfigForBizNotify(),
        },
      };

    default:
      throw new Error("Invalid value of `nodeType`");
  }
};

export const duplicateNode = (node: WorkflowNode, options?: { withCopySuffix?: boolean }) => {
  return duplicateNodes([node], options)[0];
};

export const duplicateNodes = (nodes: WorkflowNode[], options?: { withCopySuffix?: boolean }) => {
  function duplicate(node: WorkflowNode, { withCopySuffix, nodeIdMap }: { withCopySuffix: boolean; nodeIdMap: Map<string, string> }) {
    const { produce } = new Immer({ autoFreeze: false });
    return produce(node, (draft) => {
      draft.data ??= {};
      draft.id = newNodeId();
      draft.data.name = withCopySuffix ? `${draft.data?.name || ""}-copy` : `${draft.data?.name || ""}`;

      nodeIdMap.set(node.id, draft.id); // 原节点 ID 映射到新节点 ID

      if (draft.blocks) {
        draft.blocks = draft.blocks.map((block) => duplicate(block, { withCopySuffix: false, nodeIdMap }));
      }

      if (draft.data?.config) {
        switch (draft.type) {
          case WORKFLOW_NODE_TYPES.BIZ_DEPLOY:
            {
              const prevNodeId = draft.data.config.certificateOutputNodeId as string;
              if (nodeIdMap.has(prevNodeId)) {
                draft.data.config = {
                  ...draft.data.config,
                  certificateOutputNodeId: nodeIdMap.get(prevNodeId),
                };
              }
            }
            break;

          case WORKFLOW_NODE_TYPES.BRANCHBLOCK:
            {
              const stack = [] as Expr[];
              const expr = draft.data.config.expression as Expr;
              if (expr) {
                stack.push(expr);
                while (stack.length > 0) {
                  const n = stack.pop()!;
                  if ("left" in n) {
                    stack.push(n.left);
                    if ("selector" in n.left) {
                      const prevNodeId = n.left.selector.id;
                      if (nodeIdMap.has(prevNodeId)) {
                        n.left.selector.id = nodeIdMap.get(prevNodeId)!;
                      }
                    }
                  }
                  if ("right" in n) {
                    stack.push(n.right);
                  }
                }

                draft.data.config = {
                  ...draft.data.config,
                  expression: expr,
                };
              }
            }
            break;
        }
      }

      return draft;
    });
  }

  const map = new Map<string, string>();
  return nodes.map((node) => duplicate(node, { withCopySuffix: options?.withCopySuffix ?? true, nodeIdMap: map }));
};
// #endregion

// #region Expression
export enum ExprType {
  Constant = "const",
  Variant = "var",
  Comparison = "comparison",
  Logical = "logical",
  Not = "not",
}

export type ExprValue = string | number | boolean;
export type ExprValueType = "string" | "number" | "boolean";
export type ExprValueSelector = {
  id: string;
  name: string;
  type: ExprValueType;
};

export type ExprComparisonOperator = "gt" | "gte" | "lt" | "lte" | "eq" | "neq";
export type ExprLogicalOperator = "and" | "or" | "not";

export type ConstantExpr = { type: ExprType.Constant; value: string; valueType: ExprValueType };
export type VariantExpr = { type: ExprType.Variant; selector: ExprValueSelector };
export type ComparisonExpr = { type: ExprType.Comparison; operator: ExprComparisonOperator; left: Expr; right: Expr };
export type LogicalExpr = { type: ExprType.Logical; operator: ExprLogicalOperator; left: Expr; right: Expr };
export type NotExpr = { type: ExprType.Not; expr: Expr };
export type Expr = ConstantExpr | VariantExpr | ComparisonExpr | LogicalExpr | NotExpr;
// #endregion
