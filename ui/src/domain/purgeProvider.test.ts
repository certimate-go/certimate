import { afterEach, describe, expect, it } from "vitest";

import { type PluginCatalogEntry } from "@/api/plugincatalog";
import { applyPluginCatalog } from "@/domain/pluginCatalog";
import { PURGE_PROVIDERS, purgeProvidersMap } from "@/domain/provider";
import { getPurgeProvider, listPurgeProviders, normalizePurgeProviderType, purgeProviderAliases, resolvePurgeProvider } from "@/domain/purgeProvider";

const pluginEntry = (overrides: Partial<PluginCatalogEntry> = {}): PluginCatalogEntry => ({
  source: "plugin-dir",
  capability: "purge",
  providerType: PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger",
  accessProviderType: "baiducloud",
  ownsAccess: false,
  category: "",
  displayNameKey: "plugin.baiducloud-cert-purger.name",
  accessDisplayNameKey: "",
  ...overrides,
});

const resetMap = () => {
  for (const [key, val] of purgeProvidersMap) {
    if (val.source === "plugin") purgeProvidersMap.delete(key);
  }
};

afterEach(() => {
  resetMap();
  applyPluginCatalog([]);
});

describe("purgeProviderAliases", () => {
  it("maps all legacy purge ids to purger plugin ids", () => {
    expect(purgeProviderAliases.size).toBe(16);
    for (const id of Object.values(PURGE_PROVIDERS)) {
      expect(purgeProviderAliases.get(id)).toBe(`${id}-purger`);
    }
  });

  it("resolves unknown provider ids unchanged", () => {
    expect(resolvePurgeProvider("webhook")).toBe("webhook");
    expect(resolvePurgeProvider("baiducloud-cert-purger")).toBe("baiducloud-cert-purger");
  });
});

describe("normalizePurgeProviderType", () => {
  it("rewrites legacy id only when the plugin entry exists", () => {
    expect(normalizePurgeProviderType(PURGE_PROVIDERS.BAIDUCLOUD_CERT)).toBe(PURGE_PROVIDERS.BAIDUCLOUD_CERT);

    applyPluginCatalog([pluginEntry()]);
    expect(normalizePurgeProviderType(PURGE_PROVIDERS.BAIDUCLOUD_CERT)).toBe(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger");
  });
});

describe("getPurgeProvider", () => {
  it("resolves legacy id to installed plugin entry", () => {
    applyPluginCatalog([pluginEntry()]);

    const provider = getPurgeProvider(PURGE_PROVIDERS.BAIDUCLOUD_CERT);
    expect(provider?.type).toBe(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger");
    expect(provider?.source).toBe("plugin");
    expect(provider?.provider).toBe("baiducloud");
  });

  it("falls back to the static entry while the plugin is not installed", () => {
    const provider = getPurgeProvider(PURGE_PROVIDERS.ALIYUN_CAS);
    expect(provider?.type).toBe(PURGE_PROVIDERS.ALIYUN_CAS);
  });

  it("returns undefined for a migrated legacy id without the plugin installed", () => {
    expect(getPurgeProvider(PURGE_PROVIDERS.BAIDUCLOUD_CERT)).toBeUndefined();
  });

  it("returns undefined for missing plugin and unknown ids", () => {
    resetMap();
    purgeProvidersMap.delete(PURGE_PROVIDERS.AWS_ACM);
    try {
      expect(getPurgeProvider(PURGE_PROVIDERS.AWS_ACM)).toBeUndefined();
    } finally {
      purgeProvidersMap.set(PURGE_PROVIDERS.AWS_ACM, {
        type: PURGE_PROVIDERS.AWS_ACM,
        name: "provider.aws_acm",
        icon: "",
        provider: "aws",
        builtin: false,
      });
    }
    expect(getPurgeProvider("never-heard-of")).toBeUndefined();
    expect(getPurgeProvider(void 0)).toBeUndefined();
  });
});

describe("applyPluginCatalog with purge entries", () => {
  it("adds purge plugins without touching other capability maps", () => {
    applyPluginCatalog([pluginEntry()]);

    expect(purgeProvidersMap.get(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger")?.source).toBe("plugin");
  });

  it("removes plugin purge entries on refresh", () => {
    applyPluginCatalog([pluginEntry()]);
    applyPluginCatalog([]);

    expect(purgeProvidersMap.has(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger")).toBe(false);
    expect(purgeProvidersMap.has(PURGE_PROVIDERS.ALIYUN_CAS)).toBe(true);
  });

  it("does not register purge plugins as deployment providers", async () => {
    const { deploymentProvidersMap } = await import("@/domain/provider");
    applyPluginCatalog([pluginEntry()]);

    expect(deploymentProvidersMap.has(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger")).toBe(false);
  });
});

describe("listPurgeProviders", () => {
  it("hides static entries superseded by an installed plugin", () => {
    applyPluginCatalog([pluginEntry()]);

    const types = listPurgeProviders().map((provider) => provider.type);
    expect(types).toContain(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger");
    expect(types).not.toContain(PURGE_PROVIDERS.BAIDUCLOUD_CERT);
    expect(types).toContain(PURGE_PROVIDERS.ALIYUN_CAS);
  });

  it("drops a migrated legacy entry entirely after the plugin is uninstalled", () => {
    applyPluginCatalog([pluginEntry()]);
    applyPluginCatalog([]);

    const types = listPurgeProviders().map((provider) => provider.type);
    expect(types).not.toContain(PURGE_PROVIDERS.BAIDUCLOUD_CERT);
    expect(types).not.toContain(PURGE_PROVIDERS.BAIDUCLOUD_CERT + "-purger");
  });
});
