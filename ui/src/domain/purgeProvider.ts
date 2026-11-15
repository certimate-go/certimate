import { PURGE_PROVIDERS, type PurgeProvider, purgeProvidersMap } from "./provider";

export const purgeProviderAliases: ReadonlyMap<string, string> = new Map(Object.values(PURGE_PROVIDERS).map((id) => [id, `${id}-purger`] as const));

export const resolvePurgeProvider = (provider: string): string => purgeProviderAliases.get(provider) ?? provider;

export const normalizePurgeProviderType = (provider: string): string => {
  const resolved = resolvePurgeProvider(provider);
  return purgeProvidersMap.has(resolved) ? resolved : provider;
};

export const getPurgeProvider = (provider: string | null | undefined): PurgeProvider | undefined => {
  if (!provider) return undefined;
  return purgeProvidersMap.get(resolvePurgeProvider(provider)) ?? purgeProvidersMap.get(provider);
};

export const listPurgeProviders = (): PurgeProvider[] => {
  const pluginTypes = new Set<string>(
    Array.from(purgeProvidersMap.values())
      .filter((provider) => provider.source === "plugin")
      .map((provider) => provider.type)
  );
  return Array.from(purgeProvidersMap.values()).filter((provider) => provider.source === "plugin" || !pluginTypes.has(resolvePurgeProvider(provider.type)));
};
