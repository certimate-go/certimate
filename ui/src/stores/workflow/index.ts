import { produce } from "immer";
import { create } from "zustand";

import {
  type WorkflowModel,
  type WorkflowNodeConfigForStart,
  addBranch,
  addNode,
  duplicateBranch,
  duplicateNode,
  getOutputBeforeNodeId,
  removeBranch,
  removeNode,
  updateNode,
} from "@/domain/workflow";
import { get as getWorkflow, save as saveWorkflow, subscribe as subscribeWorkflow } from "@/repository/workflow";

import { type WorkflowState, type WorkflowStore } from "./types";

export const useWorkflowStore = create<WorkflowStore>((set, get) => {
  const ensureInitialized = () => {
    if (!get().initialized) throw "Workflow not initialized yet";
  };

  let unsubscriber: (() => void) | undefined = undefined;

  return {
    workflow: {} as WorkflowModel,
    initialized: false,

    init: async (id: string) => {
      const data = await getWorkflow(id);
      set({
        workflow: data,
        initialized: true,
      });

      unsubscriber ??= await subscribeWorkflow(id, (cb) => {
        if (cb.record.id !== get().workflow.id) return;

        set({
          workflow: cb.record,
        });
      });
    },

    destroy: () => {
      unsubscriber?.();
      unsubscriber = undefined;

      set({
        workflow: {} as WorkflowModel,
        initialized: false,
      });
    },

    setBaseInfo: async (name, description) => {
      ensureInitialized();

      const resp = await saveWorkflow({
        id: get().workflow.id!,
        name: name || "",
        description: description || "",
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.name = resp.name;
            draft.description = resp.description;
          }),
        };
      });
    },

    setEnabled: async (enabled) => {
      ensureInitialized();

      const resp = await saveWorkflow({
        id: get().workflow.id!,
        enabled: enabled,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.enabled = resp.enabled;
          }),
        };
      });
    },

    publish: async () => {
      ensureInitialized();

      const root = get().workflow.draft!;
      const startConfig = root.config as WorkflowNodeConfigForStart;
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        trigger: startConfig.trigger,
        triggerCron: startConfig.triggerCron,
        content: root,
        hasDraft: false,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.trigger = resp.trigger;
            draft.triggerCron = resp.triggerCron;
            draft.content = resp.content;
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    rollback: async () => {
      ensureInitialized();

      const root = get().workflow.content!;
      const startConfig = root.config as WorkflowNodeConfigForStart;
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: false,
        trigger: startConfig.trigger,
        triggerCron: startConfig.triggerCron,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.trigger = resp.trigger;
            draft.triggerCron = resp.triggerCron;
            draft.content = resp.content;
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    addNode: async (node, previousNodeId) => {
      ensureInitialized();

      const root = addNode(get().workflow.draft!, node, previousNodeId);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    duplicateNode: async (node) => {
      ensureInitialized();

      const root = duplicateNode(get().workflow.draft!, node);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    updateNode: async (node) => {
      ensureInitialized();

      const root = updateNode(get().workflow.draft!, node);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    removeNode: async (node) => {
      ensureInitialized();

      const root = removeNode(get().workflow.draft!, node.id);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    addBranch: async (branchId) => {
      ensureInitialized();

      const root = addBranch(get().workflow.draft!, branchId);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    duplicateBranch: async (branchId, index) => {
      ensureInitialized();

      const root = duplicateBranch(get().workflow.draft!, branchId, index);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    removeBranch: async (branchId, index) => {
      ensureInitialized();

      const root = removeBranch(get().workflow.draft!, branchId, index);
      const resp = await saveWorkflow({
        id: get().workflow.id!,
        draft: root,
        hasDraft: true,
      });

      set((state: WorkflowState) => {
        return {
          workflow: produce(state.workflow, (draft) => {
            draft.draft = resp.draft;
            draft.hasDraft = resp.hasDraft;
          }),
        };
      });
    },

    getWorkflowOuptutBeforeId: (nodeId, typeFilter) => {
      return getOutputBeforeNodeId(get().workflow.draft!, nodeId, typeFilter);
    },
  };
});
