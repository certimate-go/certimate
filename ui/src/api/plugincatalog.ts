import { get as httpGet } from "./_api";

export interface PluginCatalogEntry {
  source: string;
  capability: string;
  providerType: string;
  accessProviderType: string;
  ownsAccess: boolean;
  deployCategory: string;
  displayNameKey: string;
  accessDisplayNameKey: string;
  icon?: string;
  i18n?: Record<string, Record<string, string>>;
  accessUsages?: string[];
  priority?: number;
  description?: string;
  deployers?: string[];
}

export const listPluginCatalog = async (): Promise<PluginCatalogEntry[]> => {
  const resp = await httpGet<PluginCatalogEntry[]>({
    url: "/api/plugin-catalog",
  });
  return resp.data ?? [];
};
