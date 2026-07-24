import i18n from "@/i18n";

import type { PluginCatalogEntry } from "@/api/plugincatalog";
import { ACCESS_USAGES, DEPLOYMENT_CATEGORIES, accessProvidersMap, deploymentProvidersMap } from "@/domain/provider";

const PLUGIN_PLACEHOLDER_ICON =
  "data:image/svg+xml," +
  encodeURIComponent(
    "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='#ea580c' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'><path d='M7 3h4a2 2 0 0 1 2 2v0a2 2 0 0 0 2 2h2a2 2 0 0 1 2 2v4a2 2 0 0 0 2 2h0a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2h-4a2 2 0 0 0-2 2v0a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-4a2 2 0 0 1 2-2h0a2 2 0 0 0 2-2V7a2 2 0 0 1 2-2z'/></svg>"
  );

export const pluginNamespace = (providerType: string) => `plugin.${providerType}`;

const usageMap: Record<string, string> = {
  dns: ACCESS_USAGES.DNS,
  hosting: ACCESS_USAGES.HOSTING,
  notification: ACCESS_USAGES.NOTIFICATION,
  ca: ACCESS_USAGES.CA,
};

export const applyPluginCatalog = (entries: PluginCatalogEntry[]): void => {
  for (const entry of entries) {
    const accessType = entry.accessProviderType || entry.providerType;

    const usages = (entry.accessUsages || ["hosting"])
      .map((u) => usageMap[u])
      .filter(Boolean);

    if (entry.providerType && !deploymentProvidersMap.has(entry.providerType)) {
      deploymentProvidersMap.set(entry.providerType, {
        type: entry.providerType as never,
        provider: accessType as never,
        name: entry.displayNameKey || `${pluginNamespace(entry.providerType)}.name`,
        icon: entry.icon || PLUGIN_PLACEHOLDER_ICON,
        builtin: false,
        source: "plugin",
        category: (entry.deployCategory as never) || DEPLOYMENT_CATEGORIES.OTHER,
      });
    }

    if (accessType && !accessProvidersMap.has(accessType)) {
      accessProvidersMap.set(accessType, {
        type: accessType as never,
        name: entry.accessDisplayNameKey || `${pluginNamespace(accessType)}.name`,
        icon: entry.icon || PLUGIN_PLACEHOLDER_ICON,
        usages: usages as never,
        builtin: false,
        source: "plugin",
        deployers: entry.deployers,
      });
    } else if (accessType && entry.deployers) {
      const existing = accessProvidersMap.get(accessType);
      if (existing) {
        existing.deployers = entry.deployers;
      }
    }

    if (entry.i18n) {
      for (const [locale, bundle] of Object.entries(entry.i18n)) {
        i18n.addResourceBundle(locale, "translation", bundle, true, true);
      }
    }
  }
};
