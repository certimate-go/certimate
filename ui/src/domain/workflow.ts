import dayjs from "dayjs";
import { Immer, produce } from "immer";
import { nanoid } from "nanoid";

import i18n from "@/i18n";

export interface WorkflowModel extends BaseModel {
  name: string;
  description?: string;
  trigger: string;
  triggerCron?: string;
  enabled?: boolean;
  content?: WorkflowNode;
  draft?: WorkflowNode;
  hasDraft?: boolean;
  lastRunId?: string;
  lastRunStatus?: string;
  lastRunTime?: string;
}

export const WORKFLOW_TRIGGERS = Object.freeze({
  AUTO: "auto",
  MANUAL: "manual",
} as const);

export type WorkflowTriggerType = (typeof WORKFLOW_TRIGGERS)[keyof typeof WORKFLOW_TRIGGERS];

// #region Node
export enum WorkflowNodeType {
  Start = "start",
  End = "end",
  Apply = "apply",
  Upload = "upload",
  Monitor = "monitor",
  Deploy = "deploy",
  Notify = "notify",
  Branch = "branch",
  Condition = "condition",
  ExecuteResultBranch = "execute_result_branch",
  ExecuteSuccess = "execute_success",
  ExecuteFailure = "execute_failure",
  Custom = "custom",
}

const workflowNodeTypeDefaultNames: Map<WorkflowNodeType, string> = new Map([
  [WorkflowNodeType.Start, i18n.t("workflow_node.start.default_name")],
  [WorkflowNodeType.End, i18n.t("workflow_node.end.default_name")],
  [WorkflowNodeType.Apply, i18n.t("workflow_node.apply.default_name")],
  [WorkflowNodeType.Upload, i18n.t("workflow_node.upload.default_name")],
  [WorkflowNodeType.Monitor, i18n.t("workflow_node.monitor.default_name")],
  [WorkflowNodeType.Deploy, i18n.t("workflow_node.deploy.default_name")],
  [WorkflowNodeType.Notify, i18n.t("workflow_node.notify.default_name")],
  [WorkflowNodeType.Branch, i18n.t("workflow_node.branch.default_name")],
  [WorkflowNodeType.Condition, i18n.t("workflow_node.condition.default_name")],
  [WorkflowNodeType.ExecuteResultBranch, i18n.t("workflow_node.execute_result_branch.default_name")],
  [WorkflowNodeType.ExecuteSuccess, i18n.t("workflow_node.execute_success.default_name")],
  [WorkflowNodeType.ExecuteFailure, i18n.t("workflow_node.execute_failure.default_name")],
]);

const workflowNodeTypeDefaultInputs: Map<WorkflowNodeType, WorkflowNodeIO[]> = new Map([
  [WorkflowNodeType.Apply, []],
  [WorkflowNodeType.Upload, []],
  [WorkflowNodeType.Monitor, []],
  [
    WorkflowNodeType.Deploy,
    [
      {
        name: "certificate",
        type: "certificate",
        required: true,
        label: i18n.t("workflow.variables.type.certificate.label"),
      },
    ],
  ],
  [WorkflowNodeType.Notify, []],
]);

const workflowNodeTypeDefaultOutputs: Map<WorkflowNodeType, WorkflowNodeIO[]> = new Map([
  [
    WorkflowNodeType.Apply,
    [
      {
        name: "certificate",
        type: "certificate",
        required: true,
        label: i18n.t("workflow.variables.type.certificate.label"),
      },
    ],
  ],
  [
    WorkflowNodeType.Upload,
    [
      {
        name: "certificate",
        type: "certificate",
        required: true,
        label: i18n.t("workflow.variables.type.certificate.label"),
      },
    ],
  ],
  [
    WorkflowNodeType.Monitor,
    [
      {
        name: "certificate",
        type: "certificate",
        required: true,
        label: i18n.t("workflow.variables.type.certificate.label"),
      },
    ],
  ],
  [WorkflowNodeType.Deploy, []],
  [WorkflowNodeType.Notify, []],
]);

export type WorkflowNode = {
  id: string;
  name: string;
  type: WorkflowNodeType;

  config?: Record<string, unknown>;
  inputs?: WorkflowNodeIO[];
  outputs?: WorkflowNodeIO[];

  next?: WorkflowNode;
  branches?: WorkflowNode[];

  validated?: boolean;
};

export type WorkflowNodeConfigForStart = {
  trigger: string;
  triggerCron?: string;
};

export const defaultNodeConfigForStart = (): Partial<WorkflowNodeConfigForStart> => {
  return {
    trigger: WORKFLOW_TRIGGERS.AUTO,
    triggerCron: "0 0 * * *",
  };
};

export type WorkflowNodeConfigForApply = {
  domains: string;
  contactEmail: string;
  challengeType: string;
  provider: string;
  providerAccessId: string;
  providerConfig?: Record<string, unknown>;
  caProvider?: string;
  caProviderAccessId?: string;
  caProviderConfig?: Record<string, unknown>;
  keyAlgorithm: string;
  lifeTime?: string;
  acmeProfile?: string;
  nameservers?: string;
  dnsPropagationTimeout?: number;
  dnsTTL?: number;
  disableFollowCNAME?: boolean;
  disableARI?: boolean;
  skipBeforeExpiryDays: number;
};

export const defaultNodeConfigForApply = (): Partial<WorkflowNodeConfigForApply> => {
  return {
    challengeType: "dns-01",
    keyAlgorithm: "RSA2048",
    skipBeforeExpiryDays: 30,
  };
};

export type WorkflowNodeConfigForUpload = {
  certificateId: string;
  domains: string;
  certificate: string;
  privateKey: string;
};

export const defaultNodeConfigForUpload = (): Partial<WorkflowNodeConfigForUpload> => {
  return {};
};

export type WorkflowNodeConfigForMonitor = {
  host: string;
  port: number;
  domain?: string;
  requestPath?: string;
};

export const defaultNodeConfigForMonitor = (): Partial<WorkflowNodeConfigForMonitor> => {
  return {
    port: 443,
    requestPath: "/",
  };
};

export type WorkflowNodeConfigForDeploy = {
  certificate: string;
  provider: string;
  providerAccessId?: string;
  providerConfig?: Record<string, unknown>;
  skipOnLastSucceeded: boolean;
};

export const defaultNodeConfigForDeploy = (): Partial<WorkflowNodeConfigForDeploy> => {
  return {
    skipOnLastSucceeded: true,
  };
};

export type WorkflowNodeConfigForNotify = {
  subject: string;
  message: string;
  /**
   * @deprecated
   */
  channel?: string;
  provider: string;
  providerAccessId: string;
  providerConfig?: Record<string, unknown>;
  skipOnAllPrevSkipped?: boolean;
};

export const defaultNodeConfigForNotify = (): Partial<WorkflowNodeConfigForNotify> => {
  return {};
};

export type WorkflowNodeConfigForCondition = {
  expression?: Expr;
};

export const defaultNodeConfigForCondition = (): Partial<WorkflowNodeConfigForCondition> => {
  return {};
};

export type WorkflowNodeConfigForBranch = never;

export type WorkflowNodeConfigForEnd = never;

export type WorkflowNodeIO = {
  name: string;
  type: string;
  required: boolean;
  label: string;
  value?: string;
  valueSelector?: WorkflowNodeIOValueSelector;
};

export type WorkflowNodeIOValueSelector = ExprValueSelector;
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

const isBranchNode = (node: WorkflowNode) => {
  return node.type === WorkflowNodeType.Branch || node.type === WorkflowNodeType.ExecuteResultBranch;
};

type InitWorkflowOptions = {
  template?: "standard" | "certtest";
};

export const initWorkflow = (options: InitWorkflowOptions = {}): WorkflowModel => {
  const root = newNode(WorkflowNodeType.Start, {
    nodeConfig: { trigger: WORKFLOW_TRIGGERS.MANUAL },
  });

  switch (options.template) {
    case "standard":
      {
        let current = root;

        const applyNode = newNode(WorkflowNodeType.Apply, {
          nodeConfig: defaultNodeConfigForApply(),
        });
        current.next = applyNode;

        current = current.next;
        current.next = newNode(WorkflowNodeType.ExecuteResultBranch);

        current = current.next!.branches![1];
        current.next = newNode(WorkflowNodeType.Notify, {
          nodeConfig: {
            ...defaultNodeConfigForNotify(),
            subject: "[Certimate] Workflow Failure Alert!",
            message: "Your workflow run for the certificate application has failed. Please check the details.",
          } as WorkflowNodeConfigForNotify,
        });

        current = applyNode.next!.branches![0];
        current.next = newNode(WorkflowNodeType.Deploy, {
          nodeConfig: {
            ...defaultNodeConfigForDeploy(),
            certificate: `${applyNode.id}#certificate`,
          } as WorkflowNodeConfigForDeploy,
        });

        current = current.next;
        current.next = newNode(WorkflowNodeType.ExecuteResultBranch);

        current = current.next!.branches![1];
        current.next = newNode(WorkflowNodeType.Notify, {
          nodeConfig: {
            ...defaultNodeConfigForNotify(),
            subject: "[Certimate] Workflow Failure Alert!",
            message: "Your workflow run for the certificate deployment has failed. Please check the details.",
          } as WorkflowNodeConfigForNotify,
        });
      }
      break;

    case "certtest":
      {
        let current = root;

        const monitorNode = newNode(WorkflowNodeType.Monitor, {
          nodeConfig: defaultNodeConfigForMonitor(),
        });
        current.next = monitorNode;

        current = current.next;
        current.next = newNode(WorkflowNodeType.ExecuteResultBranch);

        current = current.next!.branches![1];
        current.next = newNode(WorkflowNodeType.Notify, {
          nodeConfig: {
            ...defaultNodeConfigForNotify(),
            subject: "[Certimate] Workflow Failure Alert!",
            message: "Your workflow run for the certificate monitoring has failed. Please check the details.",
          } as WorkflowNodeConfigForNotify,
        });

        current = monitorNode.next!.branches![0];
        const branchNode = newNode(WorkflowNodeType.Branch);
        current.next = branchNode;

        current = branchNode.branches![0];
        current.name = i18n.t("workflow_node.condition.default_name.template_certtest_on_expire_soon");
        current.config = {
          expression: {
            left: {
              left: {
                selector: {
                  id: monitorNode.id,
                  name: "certificate.validity",
                  type: "boolean",
                },
                type: "var",
              },
              operator: "eq",
              right: {
                type: "const",
                value: "true",
                valueType: "boolean",
              },
              type: "comparison",
            },
            operator: "and",
            right: {
              left: {
                selector: {
                  id: monitorNode.id,
                  name: "certificate.daysLeft",
                  type: "number",
                },
                type: "var",
              },
              operator: "lte",
              right: {
                type: "const",
                value: "30",
                valueType: "number",
              },
              type: "comparison",
            },
            type: "logical",
          },
        } as WorkflowNodeConfigForCondition;
        current.next = newNode(WorkflowNodeType.Notify, {
          nodeConfig: {
            ...defaultNodeConfigForNotify(),
            subject: "[Certimate] Certificate Expiry Alert!",
            message: "The certificate will expire soon. Please pay attention to your website.",
          } as WorkflowNodeConfigForNotify,
        });

        current = branchNode.branches![1];
        current.name = i18n.t("workflow_node.condition.default_name.template_certtest_on_expired");
        current.config = {
          expression: {
            left: {
              selector: {
                id: monitorNode.id,
                name: "certificate.validity",
                type: "boolean",
              },
              type: "var",
            },
            operator: "eq",
            right: {
              type: "const",
              value: "false",
              valueType: "boolean",
            },
            type: "comparison",
          },
        } as WorkflowNodeConfigForCondition;
        current.next = newNode(WorkflowNodeType.Notify, {
          nodeConfig: {
            ...defaultNodeConfigForNotify(),
            subject: "[Certimate] Certificate Expiry Alert!",
            message: "The certificate has already expired. Please pay attention to your website.",
          } as WorkflowNodeConfigForNotify,
        });
      }
      break;
  }

  return {
    id: null!,
    name: `MyWorkflow-${dayjs().format("YYYYMMDDHHmmss")}`,
    trigger: root.config!.trigger as string,
    triggerCron: root.config!.triggerCron as string,
    enabled: false,
    draft: root,
    hasDraft: true,
    created: new Date().toISOString(),
    updated: new Date().toISOString(),
  };
};

type NewNodeOptions = {
  nodeName?: string;
  nodeConfig?: Record<string, unknown>;
  branchIndex?: number;
};

export const newNode = (nodeType: WorkflowNodeType, options: NewNodeOptions = {}): WorkflowNode => {
  const nodeTypeName = workflowNodeTypeDefaultNames.get(nodeType) || "";
  const nodeName = options.branchIndex != null ? `${nodeTypeName} ${options.branchIndex + 1}` : nodeTypeName;

  const node: WorkflowNode = {
    id: nanoid(),
    name: options.nodeName ?? nodeName,
    type: nodeType,
    config: options.nodeConfig,
  };

  switch (nodeType) {
    case WorkflowNodeType.Apply:
    case WorkflowNodeType.Upload:
    case WorkflowNodeType.Monitor:
    case WorkflowNodeType.Deploy:
      {
        node.inputs = workflowNodeTypeDefaultInputs.get(nodeType);
        node.outputs = workflowNodeTypeDefaultOutputs.get(nodeType);
      }
      break;

    case WorkflowNodeType.Condition:
      {
        node.validated = true;
      }
      break;

    case WorkflowNodeType.Branch:
      {
        node.branches = [newNode(WorkflowNodeType.Condition, { branchIndex: 0 }), newNode(WorkflowNodeType.Condition, { branchIndex: 1 })];
      }
      break;

    case WorkflowNodeType.ExecuteResultBranch:
      {
        node.branches = [newNode(WorkflowNodeType.ExecuteSuccess), newNode(WorkflowNodeType.ExecuteFailure)];
      }
      break;

    case WorkflowNodeType.ExecuteSuccess:
    case WorkflowNodeType.ExecuteFailure:
      {
        node.validated = true;
      }
      break;
  }

  return node;
};

type CloneNodeOptions = {
  withCopySuffix?: boolean;
  nodeIdMap?: Map<string, string>;
};

export const cloneNode = (sourceNode: WorkflowNode, { withCopySuffix, nodeIdMap }: CloneNodeOptions = { withCopySuffix: true }): WorkflowNode => {
  const { produce } = new Immer({ autoFreeze: false });
  const deepClone = (node: WorkflowNode): WorkflowNode => {
    return produce(node, (draft) => {
      draft.id = nanoid();

      nodeIdMap ??= new Map(); // 原节点 ID 映射到新节点 ID
      nodeIdMap.set(node.id, draft.id);

      if (draft.next) {
        draft.next = cloneNode(draft.next, { withCopySuffix, nodeIdMap });
      }

      if (draft.branches) {
        draft.branches = draft.branches.map((branch) => cloneNode(branch, { withCopySuffix, nodeIdMap }));
      }

      if (draft.config) {
        switch (draft.type) {
          case WorkflowNodeType.Deploy:
            {
              const prevNodeId = (draft.config as WorkflowNodeConfigForDeploy).certificate?.split("#")?.[0];
              if (nodeIdMap.has(prevNodeId)) {
                draft.config = {
                  ...draft.config,
                  certificate: `${nodeIdMap.get(prevNodeId)}#certificate`,
                } as WorkflowNodeConfigForDeploy;
              }
            }
            break;

          case WorkflowNodeType.Condition:
            {
              const stack = [] as Expr[];

              const expr = (draft.config as WorkflowNodeConfigForCondition).expression;
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

                draft.config = {
                  ...draft.config,
                  expression: expr,
                } as WorkflowNodeConfigForCondition;
              }
            }
            break;
        }
      }

      return draft;
    });
  };

  const copyNode = produce(sourceNode, (draft) => {
    draft.name = withCopySuffix ? `${draft.name}-copy` : draft.name;
  });
  return deepClone(copyNode);
};

export const addNode = (root: WorkflowNode, targetNode: WorkflowNode, previousNodeId: string) => {
  return produce(root, (draft) => {
    let current = draft;
    while (current) {
      if (current.id === previousNodeId && !isBranchNode(targetNode)) {
        targetNode.next = current.next;
        current.next = targetNode;
        break;
      } else if (current.id === previousNodeId && isBranchNode(targetNode)) {
        targetNode.branches![0].next = current.next;
        current.next = targetNode;
        break;
      }

      if (isBranchNode(current)) {
        current.branches ??= [];
        current.branches = current.branches.map((branch) => addNode(branch, targetNode, previousNodeId));
      }

      current = current.next as WorkflowNode;
    }

    return draft;
  });
};

export const duplicateNode = (root: WorkflowNode, targetNode: WorkflowNode) => {
  if (isBranchNode(targetNode)) {
    throw new Error("Cannot duplicate a branch node directly. Use `duplicateBranch` instead.");
  }

  const copiedNode = cloneNode(targetNode);
  return addNode(root, copiedNode, targetNode.id);
};

export const updateNode = (root: WorkflowNode, targetNode: WorkflowNode) => {
  if (isBranchNode(targetNode)) {
    throw new Error("Cannot update a branch node directly. Use `updateBranch` instead.");
  }

  return produce(root, (draft) => {
    let current = draft;
    while (current) {
      if (current.id === targetNode.id) {
        // Object.assign(current, targetNode);
        // TODO: 暂时这么处理，避免 #485 #489，后续再优化
        current.type = targetNode.type;
        current.name = targetNode.name;
        current.config = targetNode.config;
        current.inputs = targetNode.inputs;
        current.outputs = targetNode.outputs;
        current.next = targetNode.next;
        current.branches = targetNode.branches;
        current.validated = targetNode.validated;
        break;
      }

      if (isBranchNode(current)) {
        current.branches ??= [];
        current.branches = current.branches.map((branch) => updateNode(branch, targetNode));
      }

      current = current.next as WorkflowNode;
    }

    return draft;
  });
};

export const removeNode = (root: WorkflowNode, targetNodeId: string) => {
  return produce(root, (draft) => {
    let current = draft;
    while (current) {
      if (current.next?.id === targetNodeId) {
        current.next = current.next.next;
        break;
      }

      if (isBranchNode(current)) {
        current.branches ??= [];
        current.branches = current.branches.map((branch) => removeNode(branch, targetNodeId));
      }

      current = current.next as WorkflowNode;
    }

    return draft;
  });
};

export const addBranch = (root: WorkflowNode, branchNodeId: string) => {
  return produce(root, (draft) => {
    let current = draft;
    while (current) {
      if (current.id === branchNodeId) {
        if (current.type !== WorkflowNodeType.Branch) {
          return draft;
        }

        current.branches ??= [];
        current.branches.push(
          newNode(WorkflowNodeType.Condition, {
            branchIndex: current.branches.length,
          })
        );
        break;
      }

      if (isBranchNode(current)) {
        current.branches ??= [];
        current.branches = current.branches.map((branch) => addBranch(branch, branchNodeId));
      }

      current = current.next as WorkflowNode;
    }

    return draft;
  });
};

export const duplicateBranch = (root: WorkflowNode, branchNodeId: string, branchIndex: number) => {
  return produce(root, (draft) => {
    let current = draft;
    let last: WorkflowNode | undefined = {
      id: "",
      name: "",
      type: WorkflowNodeType.Start,
      next: draft,
    };
    while (current && last) {
      if (current.id === branchNodeId) {
        if (!isBranchNode(current)) {
          return draft;
        }

        current.branches ??= [];
        current.branches.splice(branchIndex + 1, 0, cloneNode(current.branches[branchIndex]));

        break;
      }

      if (isBranchNode(current)) {
        current.branches ??= [];
        current.branches = current.branches.map((branch) => duplicateBranch(branch, branchNodeId, branchIndex));
      }

      current = current.next as WorkflowNode;
      last = last.next;
    }

    return draft;
  });
};

export const removeBranch = (root: WorkflowNode, branchNodeId: string, branchIndex: number) => {
  return produce(root, (draft) => {
    let current = draft;
    let last: WorkflowNode | undefined = {
      id: "",
      name: "",
      type: WorkflowNodeType.Start,
      next: draft,
    };
    while (current && last) {
      if (current.id === branchNodeId) {
        if (!isBranchNode(current)) {
          return draft;
        }

        current.branches ??= [];
        current.branches.splice(branchIndex, 1);

        // 如果仅剩一个分支，删除分支节点，将分支节点的下一个节点挂载到当前节点
        if (current.branches.length === 1) {
          const branch = current.branches[0];
          if (branch.next) {
            last.next = branch.next;
            let lastNode: WorkflowNode | undefined = branch.next;
            while (lastNode?.next) {
              lastNode = lastNode.next;
            }
            lastNode.next = current.next;
          } else {
            last.next = current.next;
          }
        }

        break;
      }

      if (isBranchNode(current)) {
        current.branches ??= [];
        current.branches = current.branches.map((branch) => removeBranch(branch, branchNodeId, branchIndex));
      }

      current = current.next as WorkflowNode;
      last = last.next;
    }

    return draft;
  });
};

export const getOutputBeforeNodeId = (root: WorkflowNode, nodeId: string, typeFilter?: string | string[]): WorkflowNode[] => {
  // 某个分支的节点，不应该能获取到相邻分支上节点的输出
  const outputs: WorkflowNode[] = [];

  const filter = (io: WorkflowNodeIO) => {
    if (typeFilter == null) {
      return true;
    }

    if (Array.isArray(typeFilter) && typeFilter.includes(io.type)) {
      return true;
    } else if (io.type === typeFilter) {
      return true;
    }

    return false;
  };

  const traverse = (current: WorkflowNode, output: WorkflowNode[]) => {
    if (!current) {
      return false;
    }
    if (current.id === nodeId) {
      return true;
    }

    if (current.type !== WorkflowNodeType.Branch && current.outputs && current.outputs.some((io) => filter(io))) {
      output.push({
        ...current,
        outputs: current.outputs.filter((io) => filter(io)),
      });
    }

    if (isBranchNode(current)) {
      let currentLength = output.length;
      const latestOutput = output.length > 0 ? output[output.length - 1] : null;
      for (const branch of current.branches!) {
        if (branch.type === WorkflowNodeType.ExecuteFailure) {
          output.splice(output.length - 1);
          currentLength -= 1;
        }
        if (traverse(branch, output)) {
          return true;
        }
        // 如果当前分支没有输出，清空之前的输出
        if (output.length > currentLength) {
          output.splice(currentLength);
        }
        if (latestOutput && branch.type === WorkflowNodeType.ExecuteFailure) {
          output.push(latestOutput);
          currentLength += 1;
        }
      }
    }

    return traverse(current.next as WorkflowNode, output);
  };

  traverse(root, outputs);
  return outputs;
};

export const isAllNodesValidated = (node: WorkflowNode): boolean => {
  let current = node as typeof node | undefined;
  while (current) {
    if (isBranchNode(current)) {
      for (const branch of current.branches!) {
        if (!isAllNodesValidated(branch)) {
          return false;
        }
      }
    } else {
      if (!current.validated) {
        return false;
      }
    }

    current = current.next;
  }

  return true;
};
