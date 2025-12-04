import { type CAProviderType } from "./provider";

export const SETTINGS_NAMES = Object.freeze({
  EMAILS: "emails",
  NOTIFY_TEMPLATE: "notifyTemplate",
  SCRIPT_TEMPLATE: "scriptTemplate",
  SSL_PROVIDER: "sslProvider",
  PERSISTENCE: "persistence",
} as const);

export type SettingsNames = (typeof SETTINGS_NAMES)[keyof typeof SETTINGS_NAMES];

export interface SettingsModel<T extends NonNullable<unknown> = any> extends BaseModel {
  name: string;
  content: T;
}

// #region Settings: Emails
export type EmailsSettingsContent = {
  emails: string[];
};
// #endregion

// #region Settings: NotifyTemplate
export type NotifyTemplateContent = {
  templates: Array<{
    name: string;
    subject: string;
    message: string;
  }>;
};
// #endregion

// #region Settings: ScriptTemplate
export type ScriptTemplateContent = {
  templates: Array<{
    name: string;
    command: string;
  }>;
};
// #endregion

// #region Settings: SSLProvider
export type SSLProviderSettingsContent = {
  provider: CAProviderType;
  config: {
    [key: string]: Record<string, unknown> | undefined;
  };
};
// #endregion

// #region Settings: Persistence
export type PersistenceSettingsContent = {
  certificatesWarningDaysBeforeExpire?: number;
  certificatesRetentionMaxDays?: number;
  workflowRunsRetentionMaxDays?: number;
};
// #endregion
