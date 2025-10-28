import { create } from "zustand";

import { getEnvironment as requestEnvironment } from "@/api/system";
import { ACCESS_PROVIDERS, accessProvidersMap } from "@/domain/provider";

import { type SystemEnvironmentStore } from "./types";

export const useSystemEnvironmentStore = create<SystemEnvironmentStore>((set, get) => ({
  environment: null,
  loadingEnvironment: false,
  loadedEnvironment: false,

  fetchEnvironment: async (refresh = true) => {
    if (!refresh && get().loadedEnvironment) {
      return get().environment;
    }
    if (get().loadingEnvironment) {
      return get().environment;
    }

    set({ loadingEnvironment: true });

    try {
      const resp = await requestEnvironment();
      const environment = resp.data ?? null;

      accessProvidersMap.get(ACCESS_PROVIDERS.DOCKERHOST)!.disabled = !(environment?.dockerHost.reachable ?? false);

      set({ environment, loadedEnvironment: true });
      return environment;
    } catch (err) {
      accessProvidersMap.get(ACCESS_PROVIDERS.DOCKERHOST)!.disabled = true;
      set({ environment: null, loadedEnvironment: false });
      return null;
    } finally {
      set({ loadingEnvironment: false });
    }
  },
}));

