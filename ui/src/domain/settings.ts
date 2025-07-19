import { type CAProviderType } from "./provider";

export const SETTINGS_NAMES = Object.freeze({
  EMAILS: "emails",
  NOTIFY_TEMPLATES: "notifyTemplates",
  /**
   * @deprecated
   */
  NOTIFY_CHANNELS: "notifyChannels",
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

// #region Settings: NotifyTemplates
export type NotifyTemplatesSettingsContent = {
  notifyTemplates: NotifyTemplate[];
};

export type NotifyTemplate = {
  subject: string;
  message: string;
};

export const defaultNotifyTemplate: NotifyTemplate = {
  subject: "有 ${COUNT} 张证书即将过期",
  message: "有 ${COUNT} 张证书即将过期，域名分别为 ${DOMAINS}，请保持关注！",
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
  workflowRunsMaxDaysRetention?: number;
  expiredCertificatesMaxDaysRetention?: number;
};
// #endregion
