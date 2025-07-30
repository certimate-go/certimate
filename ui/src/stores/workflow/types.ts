import { type WorkflowModel, type WorkflowNode } from "@/domain/workflow";

export interface WorkflowState {
  workflow: WorkflowModel;
  initialized: boolean;
}

export interface WorkflowActions {
  init(id: string): void;
  destroy(): void;

  setBaseInfo: (name: string, description: string) => void;
  setEnabled(enabled: boolean): void;
  publish(): void;
  rollback(): void;

  addNode: (node: WorkflowNode, previousNodeId: string) => void;
  duplicateNode: (node: WorkflowNode) => void;
  updateNode: (node: WorkflowNode) => void;
  removeNode: (node: WorkflowNode) => void;
  addBranch: (branchId: string) => void;
  duplicateBranch: (branchId: string, index: number) => void;
  removeBranch: (branchId: string, index: number) => void;

  getWorkflowOuptutBeforeId: (nodeId: string, typeFilter?: string | string[]) => WorkflowNode[];
}

export interface WorkflowStore extends WorkflowState, WorkflowActions {}
