import { type PersistenceSettingsContent } from "@/domain/settings";

export interface PersistenceSettingsState {
  settings: PersistenceSettingsContent;
  loading: boolean;
  loadedAtOnce: boolean;
}

export interface PersistenceSettingsActions {
  loadSettings: (refresh?: boolean) => Promise<void>;
  saveSettings: (settings: PersistenceSettingsContent) => Promise<void>;
}

export interface PersistenceSettingsStore extends PersistenceSettingsState, PersistenceSettingsActions {}
