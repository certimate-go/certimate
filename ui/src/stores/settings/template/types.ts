type NotifyTemplate = {
  name: string;
  subject: string;
  message: string;
};

export interface NotifyTemplatesState {
  templates: NotifyTemplate[];
  loading: boolean;
  loadedAtOnce: boolean;
}

export interface NotifyTemplatesActions {
  fetchTemplates: (refresh?: boolean) => Promise<void>;
  setTemplates: (templates: NotifyTemplate[]) => Promise<void>;
  addTemplate: (template: NotifyTemplate) => Promise<void>;
  removeTemplateByIndex: (index: number) => Promise<void>;
  removeTemplateByName: (name: string) => Promise<void>;
}

export interface NotifyTemplatesStore extends NotifyTemplatesState, NotifyTemplatesActions {}

type ScriptTemplate = {
  name: string;
  command: string;
};

export interface ScriptTemplatesState {
  templates: ScriptTemplate[];
  loading: boolean;
  loadedAtOnce: boolean;
}

export interface ScriptTemplatesActions {
  fetchTemplates: (refresh?: boolean) => Promise<void>;
  setTemplates: (templates: ScriptTemplate[]) => Promise<void>;
  addTemplate: (template: ScriptTemplate) => Promise<void>;
  removeTemplateByIndex: (index: number) => Promise<void>;
  removeTemplateByName: (name: string) => Promise<void>;
}

export interface ScriptTemplatesStore extends ScriptTemplatesState, ScriptTemplatesActions {}
