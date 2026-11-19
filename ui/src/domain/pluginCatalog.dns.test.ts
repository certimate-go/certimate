import { afterEach, expect, it } from "vitest";

import { applyPluginCatalog } from "@/domain/pluginCatalog";
import { acmeDns01ProvidersMap, getACMEDns01Provider, matchesACMEDns01Access } from "@/domain/provider";

afterEach(() => applyPluginCatalog([]));

it("resolves an installed DNS alias while listing only the canonical provider", () => {
  const canonical = "demo-cloud-dns";
  const alias = "demo-cloud";
  expect(getACMEDns01Provider(alias)).toBeUndefined();

  applyPluginCatalog([
    {
      source: "plugin",
      capability: "dns",
      providerType: canonical,
      providerAliases: [alias],
      accessProviderType: "cloudflare",
      ownsAccess: false,
      category: "",
      displayNameKey: "plugin.demo-cloud-dns.name",
      accessDisplayNameKey: "",
    },
  ]);

  expect(getACMEDns01Provider(alias)?.type).toBe(canonical);
  expect(matchesACMEDns01Access(alias, "cloudflare")).toBe(true);
  expect(acmeDns01ProvidersMap.has(alias)).toBe(false);

  applyPluginCatalog([]);
  expect(getACMEDns01Provider(alias)).toBeUndefined();
});
