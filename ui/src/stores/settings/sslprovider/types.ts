import { type SSLProviderSettingsContent } from "@/domain/settings";

export interface SSLProviderSettingsState {
  settings: SSLProviderSettingsContent;
  loading: boolean;
  loadedAtOnce: boolean;
}

export interface SSLProviderSettingsActions {
  loadSettings: (refresh?: boolean) => Promise<void>;
  saveSettings: (settings: SSLProviderSettingsContent) => Promise<void>;
}

export interface SSLProviderSettingsStore extends SSLProviderSettingsState, SSLProviderSettingsActions {}
