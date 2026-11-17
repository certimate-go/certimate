import { type WorkflowGraph } from "@/domain/workflow";
import { type WorkflowTemplateModel, type WorkflowTemplateParam, type WorkflowTemplateParamValue } from "@/domain/workflowTemplate";

import { del as httpDelete, get as httpGet, post as httpPost, put as httpPut } from "./_api";

export const listWorkflowTemplates = async (keyword?: string) => {
  const resp = await httpGet<WorkflowTemplateModel[]>({
    url: `/api/workflow-templates${keyword ? `?keyword=${encodeURIComponent(keyword)}` : ""}`,
  });
  return resp.data ?? [];
};

export const getWorkflowTemplate = async (id: string) => {
  const resp = await httpGet<WorkflowTemplateModel>({
    url: `/api/workflow-templates/${encodeURIComponent(id)}`,
  });
  return resp.data;
};

export interface WorkflowTemplateSaveRequest {
  name: string;
  description?: string;
  graph: WorkflowGraph;
  params: WorkflowTemplateParam[];
}

export const createWorkflowTemplate = async (req: WorkflowTemplateSaveRequest) => {
  const resp = await httpPost<WorkflowTemplateModel>({
    url: `/api/workflow-templates`,
    body: req,
  });
  return resp.data;
};

export const updateWorkflowTemplate = async (id: string, req: WorkflowTemplateSaveRequest & { expectedRevision: number }) => {
  const resp = await httpPut<WorkflowTemplateModel>({
    url: `/api/workflow-templates/${encodeURIComponent(id)}`,
    body: req,
  });
  return resp.data;
};

export const removeWorkflowTemplate = async (id: string) => {
  await httpDelete({
    url: `/api/workflow-templates/${encodeURIComponent(id)}`,
  });
};

export interface WorkflowTemplateInstantiateRequest {
  revision: number;
  name: string;
  description?: string;
  params: Record<string, WorkflowTemplateParamValue>;
}

export interface WorkflowTemplateInstantiateResponse {
  workflowId: string;
}

export const instantiateWorkflowTemplate = async (id: string, req: WorkflowTemplateInstantiateRequest) => {
  const resp = await httpPost<WorkflowTemplateInstantiateResponse>({
    url: `/api/workflow-templates/${encodeURIComponent(id)}/instantiate`,
    body: req,
  });
  return resp.data;
};
