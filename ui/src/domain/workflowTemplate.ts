import { type WorkflowGraph } from "./workflow";

export const WORKFLOW_TEMPLATE_PARAM_KINDS = Object.freeze({
  DOMAINS: "domains",
  DEPLOYMENT: "deployment",
} as const);

export type WorkflowTemplateParamKind = (typeof WORKFLOW_TEMPLATE_PARAM_KINDS)[keyof typeof WORKFLOW_TEMPLATE_PARAM_KINDS];

export interface WorkflowTemplateParam {
  id: string;
  kind: WorkflowTemplateParamKind;
  title: string;
  description?: string;
  order: number;
  applyNodeIds?: string[];
  deployNodeId?: string;
}

export interface WorkflowTemplateModel extends BaseModel {
  name: string;
  description?: string;
  formatVersion: number;
  revision: number;
  graph: WorkflowGraph;
  params: WorkflowTemplateParam[];
}

export interface WorkflowTemplateParamValue {
  domains?: string[];
  providerAccessId?: string;
  providerConfig?: Record<string, unknown>;
}

export interface WorkflowTemplateParamError {
  paramId: string;
  code?: string;
  message: string;
}

export const sortWorkflowTemplateParams = (params: WorkflowTemplateParam[]): WorkflowTemplateParam[] => {
  return [...params].sort((a, b) => a.order - b.order || a.id.localeCompare(b.id));
};

export const findWorkflowGraphNodeById = (graph: WorkflowGraph | undefined, nodeId: string | undefined) => {
  if (!graph?.nodes || !nodeId) return undefined;

  const nodes = [...graph.nodes];
  while (nodes.length > 0) {
    const node = nodes.shift()!;
    if (node.id === nodeId) return node;
    if (node.blocks?.length) nodes.push(...node.blocks);
  }

  return undefined;
};

export const listWorkflowGraphNodesByType = (graph: WorkflowGraph | undefined, nodeType: string) => {
  const results: WorkflowGraph["nodes"] = [];

  const walk = (nodes: WorkflowGraph["nodes"]) => {
    for (const node of nodes ?? []) {
      if (node.type === nodeType) results.push(node);
      if (node.blocks?.length) walk(node.blocks);
    }
  };
  walk(graph?.nodes ?? []);

  return results;
};
