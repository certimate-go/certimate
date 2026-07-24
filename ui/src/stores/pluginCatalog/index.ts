import { create } from "zustand";

import { listPluginCatalog } from "@/api/plugincatalog";
import { applyPluginCatalog } from "@/domain/pluginCatalog";

interface PluginCatalogState {
  loaded: boolean;
  init: () => Promise<void>;
}

export const usePluginCatalogStore = create<PluginCatalogState>((set) => {
  let inflight: Promise<void> | null = null;
  return {
    loaded: false,
    init: () => {
      if (inflight) {
        return inflight;
      }
      inflight = (async () => {
        try {
          const entries = await listPluginCatalog();
          applyPluginCatalog(entries);
        } catch {
          // graceful degrade: builtin providers still work
        } finally {
          set({ loaded: true });
          inflight = null;
        }
      })();
      return inflight;
    },
  };
});
