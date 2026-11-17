import { create } from "zustand";

import { listPluginCatalog } from "@/api/plugincatalog";
import { clearProviderSchemaCache } from "@/api/providerschema";
import { applyPluginCatalog } from "@/domain/pluginCatalog";

interface PluginCatalogState {
  loaded: boolean;
  version: number;
  init: () => Promise<void>;
  reload: () => Promise<void>;
}

export const usePluginCatalogStore = create<PluginCatalogState>((set) => {
  let inflight: Promise<void> | null = null;

  const run = (): Promise<void> => {
    if (inflight) {
      return inflight;
    }
    inflight = (async () => {
      try {
        const entries = await listPluginCatalog();
        applyPluginCatalog(entries);
        clearProviderSchemaCache();
        set((s) => ({ loaded: true, version: s.version + 1 }));
      } catch {
        set({ loaded: true });
      } finally {
        inflight = null;
      }
    })();
    return inflight;
  };

  return {
    loaded: false,
    version: 0,
    init: run,
    reload: async () => {
      clearProviderSchemaCache();
      if (inflight) {
        await inflight;
      }
      await run();
    },
  };
});
