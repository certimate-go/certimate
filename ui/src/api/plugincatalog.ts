import { get as httpGet } from "./_api";

export interface PluginCatalogEntry {
  source: string;
  providerType: string;
  accessProviderType: string;
  deployCategory: string;
  displayNameKey: string;
  accessDisplayNameKey: string;
  icon?: string;
  i18n?: Record<string, Record<string, string>>;
}

export const listPluginCatalog = async (): Promise<PluginCatalogEntry[]> => {
  try {
    const resp = await httpGet<PluginCatalogEntry[]>({
      url: "/api/plugin-catalog",
    });
    return resp.data ?? [];
  } catch {
    return [];
  }
};
