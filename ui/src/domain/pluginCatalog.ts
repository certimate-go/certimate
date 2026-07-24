import i18n from "@/i18n";

import type { PluginCatalogEntry } from "@/api/plugincatalog";
import { ACCESS_USAGES, DEPLOYMENT_CATEGORIES, accessProvidersMap, deploymentProvidersMap } from "@/domain/provider";

const PLUGIN_PLACEHOLDER_ICON =
  "data:image/svg+xml," +
  encodeURIComponent(
    "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='#ea580c' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><rect x='3' y='3' width='18' height='18' rx='3'/><path d='M8 8h8M8 12h8M8 16h5'/></svg>"
  );

export const pluginNamespace = (providerType: string) => `plugin.${providerType}`;

export const applyPluginCatalog = (entries: PluginCatalogEntry[]): void => {
  for (const entry of entries) {
    const accessType = entry.accessProviderType || entry.providerType;

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
        usages: [ACCESS_USAGES.HOSTING],
        builtin: false,
        source: "plugin",
      });
    }

    if (entry.i18n) {
      for (const [locale, bundle] of Object.entries(entry.i18n)) {
        i18n.addResourceBundle(locale, "translation", bundle, true, true);
      }
    }
  }
};
