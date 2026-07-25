import { create } from "zustand";

import { listPluginCatalog } from "@/api/plugincatalog";
import { applyPluginCatalog } from "@/domain/pluginCatalog";

interface PluginCatalogState {
  loaded: boolean;
  init: () => Promise<void>;
  reload: () => Promise<void>;
}

export const usePluginCatalogStore = create<PluginCatalogState>((set, get) => {
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
        } finally {
          set({ loaded: true });
          inflight = null;
        }
      })();
      return inflight;
    },
    reload: () => {
      if (!get().loaded) {
        return Promise.resolve();
      }
      if (inflight) {
        return inflight;
      }
      inflight = (async () => {
        try {
          const entries = await listPluginCatalog();
          applyPluginCatalog(entries);
        } catch {
        } finally {
          inflight = null;
        }
      })();
      return inflight;
    },
  };
});
